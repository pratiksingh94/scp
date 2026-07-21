package scp

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MessageType identifies the type of an SCP packet
type MessageType byte

// SCP message types
const (
	MsgClientHello MessageType = 0x01 // opens the handshake
	MsgServerHello MessageType = 0x02 // server response to ClientHello
	MsgDone        MessageType = 0x03 // handshake verification
	MsgData        MessageType = 0x04 // encrypted application data
	MsgError       MessageType = 0x05 // protocol error, terminates connection
)

// Packet is the basic unit of the SCP wire format
// Every SCP message is framed as a 5-byte header (type + length) followed by a payload of exactly length bytes
type Packet struct {
	Type    MessageType
	Payload []byte
}

// WritePacket serializes p and writes it to w
// The format is: [1B type][4B big-endian length][NB payload]
func WritePacket(w io.Writer, p Packet) error {
	header := make([]byte, 5)
	header[0] = byte(p.Type)

	binary.BigEndian.PutUint32(header[1:], uint32(len(p.Payload)))

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if len(p.Payload) > 0 {
		if _, err := w.Write(p.Payload); err != nil {
			return fmt.Errorf("write payload: %w", err)
		}
	}

	return nil
}

// ReadPacket reads one packet from r and returns it
// Blocks until a complete packet is available or an error occurs
func ReadPacket(r io.Reader) (Packet, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(r, header); err != nil {
		return Packet{}, fmt.Errorf("read header: %w", err)
	}

	msgType := MessageType(header[0])
	payloadLen := binary.BigEndian.Uint32(header[1:])

	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return Packet{}, fmt.Errorf("read payload: %w", err)
		}
	}

	packet := Packet{
		Type:    msgType,
		Payload: payload,
	}

	return packet, nil
}
