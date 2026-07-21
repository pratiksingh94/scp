package scp

// all the crypto helper functions

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// NonceCounter generates monotonically increasing 12-byte nonces for use with ChaCha20-Poly1305
// The counter is encoded as a big-endian uint64 in the last 8 bytes of the nonce
// NOT safe for concurrent use
type NonceCounter struct {
	counter uint64
}

// NewNonceCounter returns a new NounceCounter starting at zero
func NewNonceCounter() *NonceCounter {
	return &NonceCounter{}
}

// Next returns the next 12-byte nonce and increments the counter
// must not be called more than 2^64 times with the same key, but tbh no one is gonna try that lmao
func (n *NonceCounter) Next() []byte {
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint64(nonce[4:], n.counter)
	n.counter++

	return nonce
}

// GenerateKeypair generates a fresh pair of ephemeral X25519 keypair
// a new keypair should be generated for EVERY handshake
func GenerateKeypair() (publicKey [32]byte, privateKey [32]byte, err error) {
	curve := ecdh.X25519()

	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return [32]byte{}, [32]byte{}, fmt.Errorf("generating keypair: %w", err)
	}

	pub := priv.PublicKey()

	return [32]byte(pub.Bytes()), [32]byte(priv.Bytes()), nil
}

// SharedSecret computes the X25519 Diffie-Hellman shared secret from a local private key and a peer's public key
// both sides independently arrive at the same shared secret
func SharedSecret(privateKey [32]byte, peerPublicKey [32]byte) ([32]byte, error) {
	curve := ecdh.X25519()

	privKey, err := curve.NewPrivateKey(privateKey[:])
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to use private key: %w", err)
	}

	pubKey, err := curve.NewPublicKey(peerPublicKey[:])
	if err != nil {
		return [32]byte{}, fmt.Errorf("failde to use peer public key: %w", err)
	}

	secret, err := privKey.ECDH(pubKey)
	if err != nil {
		return [32]byte{}, fmt.Errorf("generating shared secret: %w", err)
	}

	return [32]byte(secret), nil
}

// ComputePSKProof computes an HMAC-SHA256 over the data fields, keyed by the PSK
// used to prove PSK knowledge during the handshake without revealing the PSK itself
func ComputePSKProof(psk []byte, data ...[]byte) []byte {
	mac := hmac.New(sha256.New, psk)

	for _, d := range data {
		mac.Write(d)
	}

	return mac.Sum(nil)
}

// DeriveSessionKey derives a 32-byte session key from ECDH shared secret using HKDF-SHA256
// The PSK is used as HKDF salt and both nonces are included into the info field to bind the key to this specific session
func DeriveSessionKey(sharedSecret [32]byte, psk []byte, clientNonce []byte, serverNonce []byte) ([]byte, error) {
	info := append([]byte("SCP-session"), clientNonce...)
	info = append(info, serverNonce...)

	h := hkdf.New(sha256.New, sharedSecret[:], psk, info)

	sessionKey := make([]byte, 32)
	if _, err := io.ReadFull(h, sessionKey); err != nil {
		return nil, fmt.Errorf("derive session key: %w", err)
	}

	return sessionKey, nil
}

// Encrypt encrypts plaintext using ChaCha20-Poly1305 with the given key and nonce
// The nonce must be 12 bytes as thats what ChaCha20 uses
// Returns ciphertext with a 16-byte Poly1305 auth tag appended
func Encrypt(key []byte, nonce []byte, plaintext []byte) ([]byte, error) {
	AEAD, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	return AEAD.Seal(nil, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext using ChaCha20-Poly1305 with the given key and nonce
// Returns an error if authentication fails, any tampering with ciphertext will cause the decryption to fail entirely
func Decrypt(key []byte, nonce []byte, ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return plaintext, nil
}
