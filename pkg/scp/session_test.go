package scp

import (
	"bytes"
	"net"
	"testing"
)

func TestSessionSendReceive(t *testing.T) {
	client, server := net.Pipe()
	psk := []byte("test-psk")

	var serverSess *Session
	done := make(chan struct{})

	go func() {
		serverSess, _ = ServerHandshake(server, psk)
		close(done)
	}()

	clientSess, err := ClientHandshake(client, psk)
	<-done

	if err != nil {
		t.Fatalf("handshake: %v", err)
	}

	msg := []byte("hi server did you know ransomware are cute")
	go func() {
		if err := clientSess.Send(msg); err != nil {
			t.Errorf("send: %v", err)
		}
	}()

	received, err := serverSess.Receive()
	if err != nil {
		t.Fatalf("receive: %v ", err)
	}

	if !bytes.Equal(received, msg) {
		t.Fatalf("got %q, want %q", received, msg)
	}

	reply := []byte("yes my phone password is ransomware")
	go func() {
		serverSess.Send(reply)
	}()

	receivedReply, err := clientSess.Receive()
	if err != nil {
		t.Fatalf("receive reply: %v", err)
	}

	if !bytes.Equal(receivedReply, reply) {
		t.Fatalf("got %q, want %q", receivedReply, reply)
	}
}
