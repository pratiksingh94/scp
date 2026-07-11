package scp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageType byte

const (
	MsgClientHello MessageType = 0x01
	MsgServerHello MessageType = 0x02
	MsgDone        MessageType = 0x03
	MsgData        MessageType = 0x04
	MsgError       MessageType = 0x05
)

type Packet struct {
	Type    MessageType
	Payload []byte
}

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
