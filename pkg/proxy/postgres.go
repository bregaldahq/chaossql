package proxy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// PGMessage represents a framed PostgreSQL wire protocol packet.
type PGMessage struct {
	Type         byte
	Length       int32
	Payload      []byte
	Raw          []byte
	IsStartup    bool
	IsSSLRequest bool
}

// ReadPGClientPacket decodes a packet from client to database.
func ReadPGClientPacket(r io.Reader, isStartup bool) (*PGMessage, error) {
	if isStartup {
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
		if length < 8 || length > 100000 {
			return nil, fmt.Errorf("invalid startup message length: %d", length)
		}

		payload := make([]byte, length-4)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}

		rawBuf := new(bytes.Buffer)
		_ = binary.Write(rawBuf, binary.BigEndian, length)
		rawBuf.Write(payload)

		code := binary.BigEndian.Uint32(payload[:4])
		if length == 8 && code == 80877103 {
			return &PGMessage{
				Length:       length,
				Payload:      payload,
				Raw:          rawBuf.Bytes(),
				IsSSLRequest: true,
			}, nil
		}

		return &PGMessage{
			Length:    length,
			Payload:   payload,
			Raw:       rawBuf.Bytes(),
			IsStartup: true,
		}, nil
	}

	// Standard framed message: [1 byte Type][4 bytes Length][Payload]
	var msgType [1]byte
	if _, err := io.ReadFull(r, msgType[:]); err != nil {
		return nil, err
	}

	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if length < 4 {
		return nil, fmt.Errorf("invalid postgres message length: %d", length)
	}

	payload := make([]byte, length-4)
	if len(payload) > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	rawBuf := new(bytes.Buffer)
	rawBuf.WriteByte(msgType[0])
	_ = binary.Write(rawBuf, binary.BigEndian, length)
	rawBuf.Write(payload)

	return &PGMessage{
		Type:    msgType[0],
		Length:  length,
		Payload: payload,
		Raw:     rawBuf.Bytes(),
	}, nil
}

// ReadPGServerPacket decodes a packet from database to client.
func ReadPGServerPacket(r io.Reader) (*PGMessage, error) {
	var msgType [1]byte
	if _, err := io.ReadFull(r, msgType[:]); err != nil {
		return nil, err
	}

	var length int32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if length < 4 {
		return nil, fmt.Errorf("invalid postgres server message length: %d", length)
	}

	payload := make([]byte, length-4)
	if len(payload) > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	rawBuf := new(bytes.Buffer)
	rawBuf.WriteByte(msgType[0])
	_ = binary.Write(rawBuf, binary.BigEndian, length)
	rawBuf.Write(payload)

	return &PGMessage{
		Type:    msgType[0],
		Length:  length,
		Payload: payload,
		Raw:     rawBuf.Bytes(),
	}, nil
}

// ParsePGQuery extracts SQL text from 'Q' (Simple Query) or 'P' (Parse) messages.
func ParsePGQuery(msg *PGMessage) (string, bool) {
	if msg == nil {
		return "", false
	}

	switch msg.Type {
	case 'Q':
		sql := strings.TrimRight(string(msg.Payload), "\x00")
		return sql, true

	case 'P':
		// Parse message: stmtName\0 query\0 ...
		idx := bytes.IndexByte(msg.Payload, 0)
		if idx == -1 || idx+1 >= len(msg.Payload) {
			return "", false
		}
		rest := msg.Payload[idx+1:]
		queryIdx := bytes.IndexByte(rest, 0)
		if queryIdx == -1 {
			return strings.TrimRight(string(rest), "\x00"), true
		}
		return string(rest[:queryIdx]), true
	}

	return "", false
}

// ParsePGCommandComplete parses the command tag from 'C' Backend messages.
func ParsePGCommandComplete(msg *PGMessage) (string, bool) {
	if msg == nil || msg.Type != 'C' {
		return "", false
	}
	tag := strings.TrimRight(string(msg.Payload), "\x00")
	return tag, true
}

// ParsePGReadyForQuery extracts the transaction status byte ('I', 'T', 'E') from 'Z' Backend messages.
func ParsePGReadyForQuery(msg *PGMessage) (byte, bool) {
	if msg == nil || msg.Type != 'Z' || len(msg.Payload) < 1 {
		return 0, false
	}
	return msg.Payload[0], true
}
