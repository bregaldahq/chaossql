package proxy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Standard MySQL command bytes.
const (
	MySQLComQuit        byte = 0x01
	MySQLComInitDB      byte = 0x02
	MySQLComQuery       byte = 0x03
	MySQLComStmtPrepare byte = 0x16
	MySQLComStmtExecute byte = 0x17
	MySQLComStmtClose   byte = 0x19
	MySQLComStmtReset   byte = 0x1a
)

// Standard MySQL server status flags.
const (
	MySQLServerStatusInTrans    uint16 = 0x0001
	MySQLServerStatusAutocommit uint16 = 0x0002
)

// Critical MySQL isolation and contention error codes.
const (
	MySQLErrLockWaitTimeout uint16 = 1205
	MySQLErrDeadlock        uint16 = 1213
)

// MySQLPacket represents a framed MySQL wire packet.
type MySQLPacket struct {
	Length  uint32
	SeqID   byte
	Command byte
	Payload []byte
	Raw     []byte
}

// ReadMySQLPacket reads and decodes a framed MySQL packet from the reader.
func ReadMySQLPacket(r io.Reader) (*MySQLPacket, error) {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	length := uint32(header[0]) | (uint32(header[1]) << 8) | (uint32(header[2]) << 16)
	seqID := header[3]

	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	rawBuf := new(bytes.Buffer)
	rawBuf.Write(header[:])
	rawBuf.Write(payload)

	var command byte
	if len(payload) > 0 {
		command = payload[0]
	}

	return &MySQLPacket{
		Length:  length,
		SeqID:   seqID,
		Command: command,
		Payload: payload,
		Raw:     rawBuf.Bytes(),
	}, nil
}

// ParseMySQLQuery extracts query strings from COM_QUERY or COM_STMT_PREPARE commands.
func ParseMySQLQuery(p *MySQLPacket) (string, bool) {
	if p == nil || len(p.Payload) < 2 {
		return "", false
	}
	if p.Command == MySQLComQuery || p.Command == MySQLComStmtPrepare {
		return string(p.Payload[1:]), true
	}
	return "", false
}

// ParseMySQLStmtExecute extracts the prepared statement ID from COM_STMT_EXECUTE.
func ParseMySQLStmtExecute(p *MySQLPacket) (uint32, bool) {
	if p == nil || len(p.Payload) < 5 || p.Command != MySQLComStmtExecute {
		return 0, false
	}
	stmtID := binary.LittleEndian.Uint32(p.Payload[1:5])
	return stmtID, true
}

// ParseMySQLOKPacket parses OK_Packet (0x00) and inspects transaction state.
func ParseMySQLOKPacket(p *MySQLPacket) (isOK bool, inTrans bool) {
	if p == nil || len(p.Payload) < 1 || p.Payload[0] != 0x00 {
		return false, false
	}

	// OK packet layout:
	// 0x00
	// affected_rows (lenenc int)
	// last_insert_id (lenenc int)
	// status_flags (2 bytes uint16)
	r := bytes.NewReader(p.Payload[1:])
	_, err := readLenEncInt(r)
	if err != nil {
		return true, false
	}
	_, err = readLenEncInt(r)
	if err != nil {
		return true, false
	}

	var status uint16
	if err := binary.Read(r, binary.LittleEndian, &status); err != nil {
		return true, false
	}

	inTrans = (status & MySQLServerStatusInTrans) != 0
	return true, inTrans
}

// ParseMySQLErrPacket parses ERR_Packet (0xFF) and extracts error code and message.
func ParseMySQLErrPacket(p *MySQLPacket) (isErr bool, errCode uint16, errMsg string) {
	if p == nil || len(p.Payload) < 3 || p.Payload[0] != 0xFF {
		return false, 0, ""
	}

	errCode = binary.LittleEndian.Uint16(p.Payload[1:3])
	rest := p.Payload[3:]
	if len(rest) > 0 && rest[0] == '#' && len(rest) > 6 {
		// SQL state marker '#' followed by 5 bytes state
		errMsg = string(rest[6:])
	} else {
		errMsg = string(rest)
	}

	return true, errCode, errMsg
}

// IsMySQLDeadlock checks whether the error code indicates a circular lock deadlock.
func IsMySQLDeadlock(errCode uint16) bool {
	return errCode == MySQLErrDeadlock
}

func readLenEncInt(r *bytes.Reader) (uint64, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}

	switch {
	case b < 0xFB:
		return uint64(b), nil
	case b == 0xFC:
		var val uint16
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return 0, err
		}
		return uint64(val), nil
	case b == 0xFD:
		var b3 [3]byte
		if _, err := io.ReadFull(r, b3[:]); err != nil {
			return 0, err
		}
		val := uint64(b3[0]) | (uint64(b3[1]) << 8) | (uint64(b3[2]) << 16)
		return val, nil
	case b == 0xFE:
		var val uint64
		if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
			return 0, err
		}
		return val, nil
	default:
		return 0, fmt.Errorf("invalid length-encoded integer header: 0x%02x", b)
	}
}
