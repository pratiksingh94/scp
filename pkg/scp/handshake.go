package scp

// BIG TODO: CLEAN UP THIS CODE AHHHHHHHHHHH

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
)

// Session represents an established SCP connection
// Use Send and Receive to exchange encrypted messages
type Session struct {
	Conn         net.Conn
	SessionKey   []byte
	NonceCounter NonceCounter
}

func RandomNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	_, err := rand.Read(nonce)
	return nonce, err
}

func ClientHandshake(conn net.Conn, psk []byte) (*Session, error) {
	// fmt.Println("client: generating keypair")
	pubKey, privKey, err := GenerateKeypair()
	if err != nil {
		return nil, fmt.Errorf("client handshake: %w", err)
	}

	nonce, err := RandomNonce(16)
	if err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	pskProof := ComputePSKProof(psk, nonce, pubKey[:])

	helloPayload := ClientHelloPayload{
		Nonce:     [16]byte(nonce),
		PublicKey: pubKey,
		PSKProof:  [32]byte(pskProof),
	}

	helloPacket := Packet{
		Type:    MsgClientHello,
		Payload: EncodeClientHello(helloPayload),
	}

	// fmt.Println("client: sending clienthello")
	err = WritePacket(conn, helloPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to send packet: %w", err)
	}

	// fmt.Println("client: waiting for serverhello")
	serverHelloPacket, err := ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("read server hello: %w", err)
	}
	if serverHelloPacket.Type == MsgError {
		errPayload, _ := DecodeError(serverHelloPacket.Payload)
		return nil, fmt.Errorf("server error: %s", errPayload.Message)
	}

	if serverHelloPacket.Type != MsgServerHello {
		WritePacket(conn, Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrInvalidMessage,
				Message: "expected ServerHello got something else",
			}),
		})

		return nil, fmt.Errorf("unexpected message type: %d", serverHelloPacket.Type)
	}

	// fmt.Println("client: got server hello")
	serverHello, err := DecodeServerHello(serverHelloPacket.Payload)
	if err != nil {
		return nil, fmt.Errorf("decoding ServerHello: %w", err)
	}

	expectedServerPSK := ComputePSKProof(psk, nonce, serverHello.Nonce[:], serverHello.PublicKey[:], pubKey[:])
	if !bytes.Equal(expectedServerPSK, serverHello.PSKProof[:]) {
		err := ErrorPayload{
			Code:    ErrInvalidPSK,
			Message: "invalid PSK received",
		}
		errPacket := Packet{
			Type:    MsgError,
			Payload: EncodeError(err),
		}

		WritePacket(conn, errPacket)
		// if anotherDamnError != nil {
		// 	return nil, fmt.Errorf("failed to send error: %w", anotherDamnError)
		// }

		return nil, fmt.Errorf("PSK verification failed")
	}

	sharedSecret, err := SharedSecret(privKey, serverHello.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("doing ECDH: %w", err)
	}

	sessionKey, err := DeriveSessionKey(sharedSecret, psk, nonce, serverHello.Nonce[:])
	if err != nil {
		return nil, fmt.Errorf("deriving session key: %w", err)
	}

	nonceCounter := NewNonceCounter()

	// todo: dont use compute PSK here as easy route to HMAC

	hmac := ComputePSKProof(sessionKey, []byte("client-done"))
	donePayload, err := Encrypt(sessionKey, nonceCounter.Next(), hmac)
	if err != nil {
		return nil, fmt.Errorf("encrypting: %w", err)
	}

	donePacket := Packet{
		Type:    MsgDone,
		Payload: donePayload,
	}

	// fmt.Println("client: sending client done")
	err = WritePacket(conn, donePacket)
	if err != nil {
		return nil, fmt.Errorf("failed to send Done pakcet: %w", err)
	}

	// fmt.Println("client: receving server done")
	serverDone, err := ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	if serverDone.Type == MsgError {
		errorPayload, _ := DecodeError(serverDone.Payload)
		return nil, fmt.Errorf("server error: %s", errorPayload.Message)
	}

	if serverDone.Type != MsgDone {
		WritePacket(conn, Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrInvalidMessage,
				Message: "expected server Done, got something else",
			}),
		})

		return nil, fmt.Errorf("unexpected message type: %d", serverDone.Type)
	}

	// fmt.Println("client: received server done")
	expectedServerDone := ComputePSKProof(sessionKey, []byte("server-done"))
	decryptedPayload, err := Decrypt(sessionKey, nonceCounter.Next(), serverDone.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt packet: %w", err)
	}

	if !bytes.Equal(decryptedPayload, expectedServerDone) {

		errPacket := Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrHandshakeFail,
				Message: "decrypted MsgDone didnt match expected value, could be invalid session key",
			}),
		}

		WritePacket(conn, errPacket)

		return nil, fmt.Errorf("MAC didnt match")
	}

	sess := Session{
		Conn:         conn,
		SessionKey:   sessionKey,
		NonceCounter: *NewNonceCounter(),
	}

	// fmt.Println("client: returning")
	return &sess, nil

}

func ServerHandshake(conn net.Conn, psk []byte) (*Session, error) {
	// fmt.Println("server: receving client hello")
	clientHelloPacket, err := ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("reading ClientHello: %w", err)
	}

	if clientHelloPacket.Type == MsgError {
		errPayload, _ := DecodeError(clientHelloPacket.Payload)
		return nil, fmt.Errorf("client error: %s", errPayload.Message)
	}

	if clientHelloPacket.Type != MsgClientHello {
		WritePacket(conn, Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrInvalidMessage,
				Message: "expected ClientHello, got something else",
			}),
		})

		return nil, fmt.Errorf("unexpected message type: %d", clientHelloPacket.Type)
	}

	// fmt.Println("server: got client hello")
	clientHello, err := DecodeClientHello(clientHelloPacket.Payload)
	if err != nil {
		return nil, fmt.Errorf("decoding ClientHello: %w", err)
	}

	expectedClientPSK := ComputePSKProof(psk, clientHello.Nonce[:], clientHello.PublicKey[:])

	if !bytes.Equal(clientHello.PSKProof[:], expectedClientPSK) {
		errPacket := Packet{
			Type:    MsgError,
			Payload: EncodeError(ErrorPayload{Code: ErrInvalidPSK, Message: "invalid PSK"}),
		}

		WritePacket(conn, errPacket)
		// if err != nil {
		// 	return nil, fmt.Errorf("writing packet: %w", err)
		// }

		return nil, fmt.Errorf("PSK verification failed")
	}

	// todo: geneate keypair
	pubKey, privKey, err := GenerateKeypair()
	if err != nil {
		return nil, fmt.Errorf("generating keypair: %w", err)
	}

	nonce, err := RandomNonce(16)
	if err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	pskProof := ComputePSKProof(psk, clientHello.Nonce[:], nonce, pubKey[:], clientHello.PublicKey[:])

	serverHelloPacket := Packet{
		Type: MsgServerHello,
		Payload: EncodeServerHello(ServerHelloPayload{
			Nonce:     [16]byte(nonce),
			PublicKey: pubKey,
			PSKProof:  [32]byte(pskProof),
		}),
	}

	// fmt.Println("server: sending server hello")
	err = WritePacket(conn, serverHelloPacket)
	if err != nil {
		return nil, fmt.Errorf("sending ServerHello: %w", err)
	}

	sharedSecret, err := SharedSecret(privKey, clientHello.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("ECDH failed: %w", err)
	}

	sessionKey, err := DeriveSessionKey(sharedSecret, psk, clientHello.Nonce[:], nonce)
	if err != nil {
		return nil, fmt.Errorf("deriving session key: %w", err)
	}

	nonceC := NewNonceCounter()

	// fmt.Println("server: receving client done")
	clientDonePacket, err := ReadPacket(conn)
	if err != nil {
		return nil, fmt.Errorf("reading client done: %w", err)
	}

	if clientDonePacket.Type == MsgError {
		errorPayload, _ := DecodeError(clientDonePacket.Payload)
		return nil, fmt.Errorf("client error: %s", errorPayload.Message)
	}

	if clientDonePacket.Type != MsgDone {
		WritePacket(conn, Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrInvalidMessage,
				Message: "expected client Done, got something else",
			}),
		})

		return nil, fmt.Errorf("unexpected message type: %d", clientDonePacket.Type)
	}

	// fmt.Println("server: got client done")
	clientDone, err := Decrypt(sessionKey, nonceC.Next(), clientDonePacket.Payload)

	if err != nil {
		return nil, fmt.Errorf("decrypting client done: %w", err)
	}

	expectedClientDone := ComputePSKProof(sessionKey, []byte("client-done"))

	if !bytes.Equal(clientDone, expectedClientDone) {
		errPacket := Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrHandshakeFail,
				Message: "decrypted MsgDone didnt match expected value, could be invalid session key",
			}),
		}

		WritePacket(conn, errPacket)

		return nil, fmt.Errorf("MAC didnt match")
	}

	// serverDonePayload, err := Encrypt(sessionKey, nonceC.Next(), []byte("server-done"))
	serverDonePayload, err := Encrypt(sessionKey, nonceC.Next(), ComputePSKProof(sessionKey, []byte("server-done")))
	if err != nil {
		return nil, fmt.Errorf("encrypting: %w", err)
	}
	serverDonePacket := Packet{
		Type:    MsgDone,
		Payload: serverDonePayload,
	}

	// fmt.Println("server: sending server done")
	err = WritePacket(conn, serverDonePacket)
	if err != nil {
		return nil, fmt.Errorf("sending packet: %w", err)
	}

	sess := Session{
		Conn:         conn,
		SessionKey:   sessionKey,
		NonceCounter: *NewNonceCounter(),
	}

	// fmt.Println("server: returning")
	return &sess, nil

}
