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

type NonceCounter struct {
	counter uint64
}

func NewNonceCounter() *NonceCounter {
	return &NonceCounter{}
}

func (n *NonceCounter) Next() []byte {
	nonce := make([]byte, 12)
	binary.BigEndian.PutUint64(nonce[4:], n.counter)
	n.counter++

	return nonce
}

func GenerateKeypair() (publicKey [32]byte, privateKey [32]byte, err error) {
	curve := ecdh.X25519()

	priv, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return [32]byte{}, [32]byte{}, fmt.Errorf("generating keypair: %w", err)
	}

	pub := priv.PublicKey()

	return [32]byte(pub.Bytes()), [32]byte(priv.Bytes()), nil
}

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

func ComputePSKProof(psk []byte, data ...[]byte) []byte {
	mac := hmac.New(sha256.New, psk)

	for _, d := range data {
		mac.Write(d)
	}

	return mac.Sum(nil)
}

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

func Encrypt(key []byte, nonce []byte, plaintext []byte) ([]byte, error) {
	AEAD, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	return AEAD.Seal(nil, nonce, plaintext, nil), nil
}

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
