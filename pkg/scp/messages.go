package scp

import "fmt"

const (
	NonceSize     = 16 // random per handsake
	PublicKeySize = 32 // X25519 public key
	PSKProofSize  = 32 // HMAC-SHA256 output
)

// CLIENT HELLOOOO ========

type ClientHelloPayload struct {
	Nonce     [NonceSize]byte
	PublicKey [PublicKeySize]byte
	PSKProof  [PSKProofSize]byte
}

func EncodeClientHello(p ClientHelloPayload) []byte {
	buf := make([]byte, 80)
	copy(buf[0:16], p.Nonce[:])
	copy(buf[16:48], p.PublicKey[:])
	copy(buf[48:80], p.PSKProof[:])

	return buf
}

func DecodeClientHello(payload []byte) (ClientHelloPayload, error) {
	if len(payload) > 80 {
		return ClientHelloPayload{}, fmt.Errorf("invalid ClientHello length: %d", len(payload))
	}

	return ClientHelloPayload{
		Nonce:     [16]byte(payload[0:16]),
		PublicKey: [32]byte(payload[16:48]),
		PSKProof:  [32]byte(payload[48:80]),
	}, nil
}

// SERVER HELLOOOOO ====================

type ServerHelloPayload struct {
	Nonce     [NonceSize]byte
	PublicKey [PublicKeySize]byte
	PSKProof  [PSKProofSize]byte
}

func EncodeServerHello(p ServerHelloPayload) []byte {
	buf := make([]byte, 80)
	copy(buf[0:16], p.Nonce[:])
	copy(buf[16:48], p.PublicKey[:])
	copy(buf[48:80], p.PSKProof[:])

	return buf
}

func DecodeServerHello(payload []byte) (ServerHelloPayload, error) {
	if len(payload) > 80 {
		return ServerHelloPayload{}, fmt.Errorf("invalid ServerHello length: %d", len(payload))
	}

	return ServerHelloPayload{
		Nonce:     [16]byte(payload[0:16]),
		PublicKey: [32]byte(payload[16:48]),
		PSKProof:  [32]byte(payload[48:80]),
	}, nil
}

// Data ============
type DataPayload struct {
	Nonce      [12]byte
	Ciphertext []byte
}

func EncodeDataPayload(p DataPayload) []byte {
	buf := make([]byte, 12+len(p.Ciphertext))
	copy(buf[0:12], p.Nonce[:])
	copy(buf[12:], p.Ciphertext)

	return buf
}

func DecodeDataPayload(payload []byte) (DataPayload, error) {
	if len(payload) < 12 {
		return DataPayload{}, fmt.Errorf("data payload too short")
	}

	return DataPayload{
		Nonce:      [12]byte(payload[0:12]),
		Ciphertext: payload[12:],
	}, nil
}
