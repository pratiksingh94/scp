package scp

import (
	"net"
	"testing"
)

func TestHandshake(t *testing.T) {
	client, server := net.Pipe()
	psk := []byte("test-psk")

	var serverSess *Session
	var serverErr error

	done := make(chan struct{})

	go func() {
		serverSess, serverErr = ServerHandshake(server, psk)
		// fmt.Println("server: goroutine done closing channel")
		close(done)
	}()

	clientSess, clientErr := ClientHandshake(client, psk)
	// fmt.Println("client: handhsake returned")
	<-done
	// fmt.Println("test: channel received")

	if clientErr != nil {
		t.Fatalf("client handhsake: %v", clientErr)
	}

	if serverErr != nil {
		t.Fatalf("server handshake: %v", serverErr)
	}

	if clientSess == nil || serverSess == nil {
		t.Fatal("nil session returned")
	}
}
