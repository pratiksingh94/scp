package scp

import (
	"bytes"
	"testing"
)

func TestReadWritePacket_RT(t *testing.T) {
	original := Packet{
		Type:    MsgData,
		Payload: []byte("we NOT using tls in big 26"),
	}

	var buf bytes.Buffer

	if err := WritePacket(&buf, original); err != nil {
		t.Fatalf("WritePacket failed: %v", err)
	}

	got, err := ReadPacket(&buf)
	if err != nil {
		t.Fatalf("ReadPcket failed: %v", err)
	}

	if got.Type != original.Type {
		t.Errorf("type mismatch, got %v original %v", got.Type, original.Type)
	}

	if !bytes.Equal(original.Payload, got.Payload) {
		t.Errorf("payload mismatch: got %v original %v", got.Payload, original.Payload)
	}
}
