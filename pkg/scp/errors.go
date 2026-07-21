package scp

import (
	"fmt"
)

// SCP error codes sent in MsgError packets
const (
	ErrUnknown        byte = 0x00 // unspecified error
	ErrInvalidPSK     byte = 0x01 // PSK proof verification failed
	ErrHandshakeFail  byte = 0x02 // Done MAC verification failed
	ErrInvalidMessage byte = 0x03 // unexpected message type received
)

type ErrorPayload struct {
	Code    byte
	Message string
}

func EncodeError(p ErrorPayload) []byte {
	msg := []byte(p.Message)
	buf := make([]byte, 1+len(msg))
	buf[0] = p.Code
	copy(buf[1:], msg)

	return buf
}

func DecodeError(payload []byte) (ErrorPayload, error) {
	if len(payload) < 1 {
		return ErrorPayload{}, fmt.Errorf("error payload too short")
	}

	return ErrorPayload{
		Code:    payload[0],
		Message: string(payload[1:]),
	}, nil
}
