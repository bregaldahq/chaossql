package proxy

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func buildMySQLPacket(seqID byte, payload []byte) []byte {
	buf := new(bytes.Buffer)
	length := len(payload)
	// 3-byte payload length (little-endian)
	buf.WriteByte(byte(length & 0xFF))
	buf.WriteByte(byte((length >> 8) & 0xFF))
	buf.WriteByte(byte((length >> 16) & 0xFF))
	buf.WriteByte(seqID)
	buf.Write(payload)
	return buf.Bytes()
}

func TestMySQL_ReadQuery(t *testing.T) {
	sql := "SELECT balance FROM accounts WHERE id = 1"
	payload := append([]byte{0x03}, []byte(sql)...) // COM_QUERY
	packet := buildMySQLPacket(0, payload)

	msg, err := ReadMySQLPacket(bytes.NewReader(packet))
	if err != nil {
		t.Fatalf("failed to read MySQL packet: %v", err)
	}

	if msg.Command != MySQLComQuery {
		t.Fatalf("expected command COM_QUERY, got 0x%02x", msg.Command)
	}

	extractedSQL, ok := ParseMySQLQuery(msg)
	if !ok || extractedSQL != sql {
		t.Fatalf("expected %q, got %q", sql, extractedSQL)
	}
}

func TestMySQL_ReadStmtPrepare(t *testing.T) {
	sql := "UPDATE accounts SET balance = balance - 50 WHERE id = 2"
	payload := append([]byte{0x16}, []byte(sql)...) // COM_STMT_PREPARE
	packet := buildMySQLPacket(0, payload)

	msg, err := ReadMySQLPacket(bytes.NewReader(packet))
	if err != nil {
		t.Fatalf("failed to read MySQL packet: %v", err)
	}

	if msg.Command != MySQLComStmtPrepare {
		t.Fatalf("expected COM_STMT_PREPARE, got 0x%02x", msg.Command)
	}

	extractedSQL, ok := ParseMySQLQuery(msg)
	if !ok || extractedSQL != sql {
		t.Fatalf("expected %q, got %q", sql, extractedSQL)
	}
}

func TestMySQL_ReadStmtExecute(t *testing.T) {
	// COM_STMT_EXECUTE: 0x17 followed by stmtID (4 bytes uint32)
	payload := []byte{0x17, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}
	packet := buildMySQLPacket(0, payload)

	msg, err := ReadMySQLPacket(bytes.NewReader(packet))
	if err != nil {
		t.Fatalf("failed to read execute packet: %v", err)
	}

	if msg.Command != MySQLComStmtExecute {
		t.Fatalf("expected COM_STMT_EXECUTE, got 0x%02x", msg.Command)
	}

	stmtID, ok := ParseMySQLStmtExecute(msg)
	if !ok || stmtID != 1 {
		t.Fatalf("expected stmtID 1, got %d", stmtID)
	}
}

func TestMySQL_ReadOKAndERRPackets(t *testing.T) {
	// 1. OK Packet (0x00) with SERVER_STATUS_IN_TRANS (0x0001)
	okPayload := []byte{0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00} // status = 0x0001
	packet := buildMySQLPacket(1, okPayload)

	msg, err := ReadMySQLPacket(bytes.NewReader(packet))
	if err != nil {
		t.Fatalf("failed to read OK packet: %v", err)
	}

	isOK, inTrans := ParseMySQLOKPacket(msg)
	if !isOK {
		t.Fatalf("expected OK packet")
	}
	if !inTrans {
		t.Fatalf("expected inTrans=true")
	}

	// 2. ERR Packet with deadlock code 1213 (0x04BD)
	errPayload := new(bytes.Buffer)
	errPayload.WriteByte(0xFF)
	_ = binary.Write(errPayload, binary.LittleEndian, uint16(1213)) // ER_LOCK_DEADLOCK
	errPayload.WriteString("#40001Deadlock found when trying to get lock; try restarting transaction")

	errPacket := buildMySQLPacket(1, errPayload.Bytes())
	msg, err = ReadMySQLPacket(bytes.NewReader(errPacket))
	if err != nil {
		t.Fatalf("failed to read ERR packet: %v", err)
	}

	isErr, errCode, errMsg := ParseMySQLErrPacket(msg)
	if !isErr {
		t.Fatalf("expected ERR packet")
	}
	if errCode != 1213 {
		t.Fatalf("expected error code 1213, got %d", errCode)
	}
	if isDeadlock := IsMySQLDeadlock(errCode); !isDeadlock {
		t.Fatalf("expected deadlock detection true")
	}
	if len(errMsg) == 0 {
		t.Fatalf("expected non-empty error message")
	}
}
