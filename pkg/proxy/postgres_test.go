package proxy

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func buildPGPacket(msgType byte, payload []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(msgType)
	length := int32(len(payload) + 4)
	_ = binary.Write(buf, binary.BigEndian, length)
	buf.Write(payload)
	return buf.Bytes()
}

func buildPGStartupPacket(user, database string) []byte {
	buf := new(bytes.Buffer)
	payload := new(bytes.Buffer)
	_ = binary.Write(payload, binary.BigEndian, int32(196608)) // Protocol 3.0
	payload.WriteString("user\x00")
	payload.WriteString(user + "\x00")
	payload.WriteString("database\x00")
	payload.WriteString(database + "\x00")
	payload.WriteByte(0) // terminating null

	length := int32(payload.Len() + 4)
	_ = binary.Write(buf, binary.BigEndian, length)
	buf.Write(payload.Bytes())
	return buf.Bytes()
}

func buildPGSSLRequest() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, int32(8))
	_ = binary.Write(buf, binary.BigEndian, int32(80877103))
	return buf.Bytes()
}

func TestPostgres_ReadStartupAndSSL(t *testing.T) {
	// 1. SSLRequest
	sslBytes := buildPGSSLRequest()
	msg, err := ReadPGClientPacket(bytes.NewReader(sslBytes), true)
	if err != nil {
		t.Fatalf("unexpected error reading SSLRequest: %v", err)
	}
	if !msg.IsSSLRequest {
		t.Fatalf("expected IsSSLRequest=true")
	}

	// 2. StartupMessage
	startupBytes := buildPGStartupPacket("postgres", "testdb")
	msg, err = ReadPGClientPacket(bytes.NewReader(startupBytes), true)
	if err != nil {
		t.Fatalf("unexpected error reading StartupMessage: %v", err)
	}
	if !msg.IsStartup {
		t.Fatalf("expected IsStartup=true")
	}
	if len(msg.Raw) != len(startupBytes) {
		t.Fatalf("expected raw length %d, got %d", len(startupBytes), len(msg.Raw))
	}
}

func TestPostgres_ReadSimpleQuery(t *testing.T) {
	sql := "SELECT id, balance FROM accounts WHERE id = 1"
	payload := append([]byte(sql), 0)
	packet := buildPGPacket('Q', payload)

	msg, err := ReadPGClientPacket(bytes.NewReader(packet), false)
	if err != nil {
		t.Fatalf("unexpected error reading SimpleQuery: %v", err)
	}
	if msg.Type != 'Q' {
		t.Fatalf("expected message type 'Q', got '%c'", msg.Type)
	}

	parsedSQL, ok := ParsePGQuery(msg)
	if !ok {
		t.Fatalf("failed to parse SQL query from message")
	}
	if parsedSQL != sql {
		t.Fatalf("expected query %q, got %q", sql, parsedSQL)
	}
}

func TestPostgres_ParseExecute(t *testing.T) {
	// 'P' Parse message: stmtName\0 query\0 numParams(int16)
	buf := new(bytes.Buffer)
	buf.WriteString("stmt1\x00")
	buf.WriteString("UPDATE accounts SET balance = balance - 100 WHERE id = 1\x00")
	_ = binary.Write(buf, binary.BigEndian, int16(0))

	packet := buildPGPacket('P', buf.Bytes())
	msg, err := ReadPGClientPacket(bytes.NewReader(packet), false)
	if err != nil {
		t.Fatalf("unexpected error reading Parse message: %v", err)
	}
	if msg.Type != 'P' {
		t.Fatalf("expected type 'P', got '%c'", msg.Type)
	}

	sql, ok := ParsePGQuery(msg)
	if !ok {
		t.Fatalf("failed to extract SQL from Parse message")
	}
	expected := "UPDATE accounts SET balance = balance - 100 WHERE id = 1"
	if sql != expected {
		t.Fatalf("expected %q, got %q", expected, sql)
	}
}

func TestPostgres_ServerMessages(t *testing.T) {
	// 1. ReadyForQuery 'Z'
	readyPacket := buildPGPacket('Z', []byte{'T'})
	msg, err := ReadPGServerPacket(bytes.NewReader(readyPacket))
	if err != nil {
		t.Fatalf("unexpected error reading ReadyForQuery: %v", err)
	}
	status, ok := ParsePGReadyForQuery(msg)
	if !ok || status != 'T' {
		t.Fatalf("expected status 'T', got '%c'", status)
	}

	// 2. CommandComplete 'C'
	completePacket := buildPGPacket('C', []byte("UPDATE 1\x00"))
	msg, err = ReadPGServerPacket(bytes.NewReader(completePacket))
	if err != nil {
		t.Fatalf("unexpected error reading CommandComplete: %v", err)
	}
	tag, ok := ParsePGCommandComplete(msg)
	if !ok || tag != "UPDATE 1" {
		t.Fatalf("expected tag 'UPDATE 1', got %q", tag)
	}
}
