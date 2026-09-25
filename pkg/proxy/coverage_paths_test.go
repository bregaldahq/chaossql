package proxy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestReadLenEncInt(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want uint64
	}{
		{"one byte", []byte{0xFA}, 250},
		{"two bytes", []byte{0xFC, 0x34, 0x12}, 0x1234},
		{"three bytes", []byte{0xFD, 0x03, 0x02, 0x01}, 0x010203},
		{"eight bytes", []byte{0xFE, 1, 0, 0, 0, 0, 0, 0, 1}, 1<<56 | 1},
	}
	for _, tc := range cases {
		got, err := readLenEncInt(bytes.NewReader(tc.in))
		if err != nil || got != tc.want {
			t.Errorf("%s: readLenEncInt = %d, %v; want %d", tc.name, got, err, tc.want)
		}
	}
	for _, bad := range [][]byte{{}, {0xFB}, {0xFF}, {0xFC, 0x01}, {0xFD, 0x01}, {0xFE, 0x01}} {
		if _, err := readLenEncInt(bytes.NewReader(bad)); err == nil {
			t.Errorf("readLenEncInt(% x) must fail", bad)
		}
	}
}

func TestParseMySQLPacketsRejectMalformedInput(t *testing.T) {
	if _, ok := ParseMySQLQuery(nil); ok {
		t.Fatal("nil packet must not parse as a query")
	}
	if _, ok := ParseMySQLQuery(&MySQLPacket{Command: 0x01, Payload: []byte{0x01, 'x'}}); ok {
		t.Fatal("COM_QUIT must not parse as a query")
	}
	if _, ok := ParseMySQLStmtExecute(&MySQLPacket{Command: MySQLComStmtExecute, Payload: []byte{0x17, 1}}); ok {
		t.Fatal("truncated COM_STMT_EXECUTE must be rejected")
	}

	ok := func(payload ...byte) *MySQLPacket { return &MySQLPacket{Payload: append([]byte{0x00}, payload...)} }
	if isOK, inTx := ParseMySQLOKPacket(ok(0x01, 0x00, byte(MySQLServerStatusInTrans), 0x00)); !isOK || !inTx {
		t.Fatalf("OK packet with IN_TRANS status = %t, %t", isOK, inTx)
	}
	for name, p := range map[string]*MySQLPacket{
		"missing affected rows":  ok(),
		"missing last insert id": ok(0x01),
		"missing status flags":   ok(0x01, 0x00, 0x01),
	} {
		if isOK, inTx := ParseMySQLOKPacket(p); !isOK || inTx {
			t.Errorf("%s: got isOK=%t inTx=%t, want an OK packet without transaction state", name, isOK, inTx)
		}
	}
}

func TestParsePGMessagesRejectOtherTypes(t *testing.T) {
	if _, ok := ParsePGQuery(nil); ok {
		t.Fatal("nil message must not parse")
	}
	if _, ok := ParsePGQuery(&PGMessage{Type: 'P', Payload: []byte("stmt\x00")}); ok {
		t.Fatal("Parse message without a query must be rejected")
	}
	if sql, ok := ParsePGQuery(&PGMessage{Type: 'P', Payload: []byte("\x00SELECT 1")}); !ok || sql != "SELECT 1" {
		t.Fatalf("unterminated Parse query = %q, %t", sql, ok)
	}
	if _, ok := ParsePGQuery(&PGMessage{Type: 'X'}); ok {
		t.Fatal("Terminate message must not parse as a query")
	}
	if _, ok := ParsePGCommandComplete(&PGMessage{Type: 'Z'}); ok {
		t.Fatal("ReadyForQuery must not parse as CommandComplete")
	}
	if _, ok := ParsePGReadyForQuery(&PGMessage{Type: 'Z'}); ok {
		t.Fatal("empty ReadyForQuery must be rejected")
	}
	for _, raw := range [][]byte{{'Q', 0, 0, 0, 3}, {'Q', 0, 0}} {
		if _, err := ReadPGServerPacket(bytes.NewReader(raw)); err == nil {
			t.Errorf("ReadPGServerPacket(% x) must fail", raw)
		}
	}
}

func TestParseIsolationLevel(t *testing.T) {
	cases := map[string]domain.IsolationLevel{
		"SET TRANSACTION ISOLATION LEVEL SERIALIZABLE":     domain.LevelSerializable,
		"SET TRANSACTION ISOLATION LEVEL REPEATABLE READ":  domain.LevelRepeatableRead,
		"SET TRANSACTION ISOLATION LEVEL READ COMMITTED":   domain.LevelReadCommitted,
		"SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED": domain.LevelReadUncommitted,
		"BEGIN": domain.LevelReadCommitted,
	}
	for sql, want := range cases {
		if got := parseIsolationLevel(sql); got != want {
			t.Errorf("parseIsolationLevel(%q) = %q, want %q", sql, got, want)
		}
	}
}

func TestNewPCTSchedulerClampsInvalidBounds(t *testing.T) {
	s := NewPCTScheduler(-time.Millisecond, -2*time.Millisecond, -time.Millisecond, -2*time.Millisecond, 0, 1)
	if s.minJitter != 0 || s.maxJitter != 0 || s.commitBarrierMin != 0 || s.commitBarrierMax != 0 || s.depth != 1 {
		t.Fatalf("scheduler = %+v, want non-negative bounds and depth 1", s)
	}
}

func TestNewServerDefaults(t *testing.T) {
	s, err := NewServer(ProxyConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if s.cfg.Protocol != ProtocolPostgres || s.cfg.ListenAddr != "127.0.0.1:5433" || s.cfg.UpstreamAddr != "127.0.0.1:5432" {
		t.Fatalf("defaults = %+v", s.cfg)
	}
}

func TestProxySessionSnapshotIsACopy(t *testing.T) {
	s := NewProxySession(1, nil, nil)
	s.SetIsolation(domain.LevelSerializable)
	e := NewShadowGraphEngine()
	e.StartTx(s, domain.LevelSerializable)
	e.RecordRead(s, []string{"accounts:1"}, "SELECT * FROM accounts WHERE id = 1")
	e.RecordWrite(s, []string{"accounts:2"}, "UPDATE accounts SET v = 1 WHERE id = 2")

	inTx, txID, reads, writes, stmts := s.GetSnapshot()
	if !inTx || txID == "" || len(reads) != 1 || len(writes) != 1 || len(stmts) != 2 {
		t.Fatalf("snapshot = %t %q %v %v %v", inTx, txID, reads, writes, stmts)
	}
	reads["accounts:99"] = struct{}{}
	if _, _, again, _, _ := s.GetSnapshot(); len(again) != 1 {
		t.Fatal("mutating the snapshot must not change the session")
	}
	if s.Isolation != domain.LevelSerializable {
		t.Fatalf("Isolation = %q", s.Isolation)
	}
}

func TestShadowGraphCallbackTraceAndGraph(t *testing.T) {
	e := NewShadowGraphEngine()
	var alerts []DetectedAnomaly
	e.SetAnomalyCallback(func(a DetectedAnomaly) { alerts = append(alerts, a) })

	s1, s2 := NewProxySession(1, nil, nil), NewProxySession(2, nil, nil)
	e.StartTx(s1, domain.LevelReadCommitted)
	e.StartTx(s2, domain.LevelReadCommitted)
	e.RecordRead(s1, []string{"accounts:1"}, "SELECT balance FROM accounts WHERE id = 1")
	e.RecordRead(s2, []string{"accounts:1"}, "SELECT balance FROM accounts WHERE id = 1")
	e.RecordWrite(s2, []string{"accounts:1"}, "UPDATE accounts SET balance = 1 WHERE id = 1")
	e.RecordCommit(s2)
	e.RecordWrite(s1, []string{"accounts:1"}, "UPDATE accounts SET balance = 2 WHERE id = 1")
	e.RecordCommit(s1)

	if len(alerts) == 0 || alerts[0].Type != domain.AnomalyLostUpdate {
		t.Fatalf("alerts = %+v, want a lost update callback", alerts)
	}
	if trace := e.GetTrace(); len(trace) == 0 {
		t.Fatal("expected recorded trace events")
	}
	if g := e.GetGraph(); len(g.Nodes) < 2 {
		t.Fatalf("graph nodes = %v, want both transactions", g.Nodes)
	}
}

func TestHandlePGHandshakeRelaysSSLNegotiation(t *testing.T) {
	clientSide, proxyClient := net.Pipe()
	proxyServer, serverSide := net.Pipe()
	defer clientSide.Close()
	defer serverSide.Close()
	session := NewProxySession(1, proxyClient, proxyServer)
	s, _ := NewServer(ProxyConfig{})

	done := make(chan error, 1)
	go func() { done <- s.handlePGHandshake(session) }()

	ssl := make([]byte, 8)
	binary.BigEndian.PutUint32(ssl[0:4], 8)
	binary.BigEndian.PutUint32(ssl[4:8], 80877103)
	go func() { _, _ = clientSide.Write(ssl) }()
	got := make([]byte, 8)
	if _, err := io.ReadFull(serverSide, got); err != nil || !bytes.Equal(got, ssl) {
		t.Fatalf("upstream received % x, %v; want the SSLRequest", got, err)
	}
	go func() { _, _ = serverSide.Write([]byte{'N'}) }()
	reply := make([]byte, 1)
	if _, err := io.ReadFull(clientSide, reply); err != nil || reply[0] != 'N' {
		t.Fatalf("client received %q, %v; want the upstream SSL refusal", reply, err)
	}

	startup := new(bytes.Buffer)
	payload := []byte{0, 3, 0, 0}
	payload = append(payload, []byte("user\x00app\x00\x00")...)
	_ = binary.Write(startup, binary.BigEndian, int32(len(payload)+4))
	startup.Write(payload)
	go func() { _, _ = clientSide.Write(startup.Bytes()) }()
	forwarded := make([]byte, startup.Len())
	if _, err := io.ReadFull(serverSide, forwarded); err != nil || !bytes.Equal(forwarded, startup.Bytes()) {
		t.Fatalf("upstream received % x, %v; want the StartupMessage", forwarded, err)
	}
	if err := <-done; err != nil {
		t.Fatalf("handshake err = %v", err)
	}
}

func TestExportSARIFAndRuleMapping(t *testing.T) {
	rules := map[domain.AnomalyType]string{
		domain.AnomalyWriteSkew:           "chaossql/A5B-write-skew",
		domain.AnomalyA5AReadSkew:         "chaossql/A5A-read-skew",
		domain.AnomalyG0DirtyWrite:        "chaossql/G0-dirty-write",
		domain.AnomalyG1aDirtyRead:        "chaossql/G1a-dirty-read",
		domain.AnomalyG1bIntermediateRead: "chaossql/G1b-intermediate-read",
		domain.AnomalyG1cCircularInfo:     "chaossql/G1c-circular-info",
		domain.AnomalyG2AntiDependency:    "chaossql/G2-anti-dependency",
		domain.AnomalyPhantom:             "chaossql/A3-phantom-read",
	}
	for anomaly, want := range rules {
		if got := mapProxyAnomalyToRuleID(anomaly); got != want {
			t.Errorf("mapProxyAnomalyToRuleID(%q) = %q, want %q", anomaly, got, want)
		}
	}

	path := filepath.Join(t.TempDir(), "proxy.sarif")
	anomalies := []DetectedAnomaly{{ID: "a1", Type: domain.AnomalyWriteSkew, TxIDs: []string{"S1-tx1"}, Items: []string{"doctors:1"}, Description: "write skew"}}
	if err := ExportSARIFToFile(anomalies, "session", path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !json.Valid(data) || !strings.Contains(string(data), "chaossql/A5B-write-skew") {
		t.Fatalf("SARIF file = %s, err = %v", data, err)
	}
	if err := ExportSARIFToFile(anomalies, "session", filepath.Join(path, "nested")); err == nil {
		t.Fatal("expected an error for an unwritable path")
	}
}

func TestDashboardRendersAnomaliesAndJSONEndpoints(t *testing.T) {
	anomaly := DetectedAnomaly{
		Type:        domain.AnomalyLostUpdate,
		TxIDs:       []string{"S1-tx1", "S2-tx1"},
		Items:       []string{"accounts:1"},
		Cycle:       []analyzer.Edge{{From: "S1-tx1", To: "S2-tx1", Type: analyzer.DepWW, Item: "accounts:1"}},
		DetectedAt:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Description: "lost update on accounts:1",
	}
	html := renderProxyDashboardHTML(ProxyConfig{}, ProxyStats{AnomaliesDetected: 1}, []DetectedAnomaly{anomaly})
	for _, want := range []string{"lost update on accounts:1", "S1-tx1, S2-tx1", "<pre>", "var(--danger)"} {
		if !strings.Contains(html, want) {
			t.Errorf("dashboard HTML missing %q", want)
		}
	}

	s, _ := NewServer(ProxyConfig{})
	httpServer, uiURL, err := StartLiveUI("127.0.0.1:0", s)
	if err != nil {
		t.Fatal(err)
	}
	defer httpServer.Close()
	for _, path := range []string{"/api/anomalies", "/api/trace"} {
		resp, err := http.Get(uiURL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "application/json" || !json.Valid(body) {
			t.Fatalf("%s: status %d, type %q, body %s", path, resp.StatusCode, resp.Header.Get("Content-Type"), body)
		}
	}
	if _, _, err := StartLiveUI("256.0.0.1:0", s); err == nil {
		t.Fatal("an invalid listen address must be rejected")
	}
}
