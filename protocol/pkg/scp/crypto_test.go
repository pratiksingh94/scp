package scp

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub1, priv1, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("unexpected bs: %v", err)
	}

	if pub1 == ([32]byte{}) {
		t.Fatal("pub key is zero")
	}
	if priv1 == ([32]byte{}) {
		t.Fatal("priv key zero")
	}

	// fmt.Println(pub1, priv1)

	pub2, priv2, _ := GenerateKeypair()
	if pub1 == pub2 {
		t.Fatal("two keypairs got same pub key")
	}
	if priv1 == priv2 {
		t.Fatal("two keypair got same priv key")
	}
}

func TestComputePSK(t *testing.T) {
	psk := []byte("test-psk")
	data1 := []byte("hello")
	data2 := []byte("nononoonoononooooononononononondont")

	proof := ComputePSKProof(psk, data1, data2)

	if len(proof) != 32 {
		t.Fatalf("erm what the sigma expected 32 got %d", len(proof))
	}

	proof2 := ComputePSKProof(psk, data1, data2)
	if !bytes.Equal(proof, proof2) {
		t.Fatal("psk proof not deterministic")
	}

	reverseOrder := ComputePSKProof(psk, data2, data1)
	if bytes.Equal(proof, reverseOrder) {
		t.Fatal("reverse order data produced same PSK")
	}

	// fmt.Println(proof)
}

func TestDeriveSessionKey(t *testing.T) {
	secret := [32]byte{}
	copy(secret[:], bytes.Repeat([]byte{0x42}, 32))

	psk := []byte("test-psk")
	clientNonce := bytes.Repeat([]byte{0x01}, 32)
	serverNonce := bytes.Repeat([]byte{0x02}, 32)

	key1, err := DeriveSessionKey(secret, psk, clientNonce, serverNonce)
	if err != nil {
		t.Fatalf("uhhh err %v", err)
	}

	key2, _ := DeriveSessionKey(secret, psk, clientNonce, serverNonce)
	if !bytes.Equal(key1, key2) {
		t.Fatal("not deterministic")
	}

	keyDiffPSK, _ := DeriveSessionKey(secret, []byte("hewwo"), clientNonce, serverNonce)
	if bytes.Equal(key1, keyDiffPSK) {
		t.Fatal("got same key from diff psk")
	}

	keyDiffClientNonce, _ := DeriveSessionKey(secret, psk, bytes.Repeat([]byte{0x99}, 32), serverNonce)
	if bytes.Equal(key1, keyDiffClientNonce) {
		t.Fatal("got same key from diff client nonce")
	}

	keyDiffServerNonce, _ := DeriveSessionKey(secret, psk, clientNonce, bytes.Repeat([]byte{0x98}, 32))
	if bytes.Equal(key1, keyDiffServerNonce) {
		t.Fatal("got sam ekey from diff server nonce")
	}

	otherSecret := [32]byte{}
	copy(otherSecret[:], bytes.Repeat([]byte{0x11}, 32))
	keyDiffSecret, _ := DeriveSessionKey(otherSecret, psk, clientNonce, serverNonce)
	if bytes.Equal(key1, keyDiffSecret) {
		t.Fatal("got same key from diff secret")
	}
}

func TestFullKeyExchange(t *testing.T) {
	//  the whole damn thang

	psk := []byte("shared-secret")
	clientNonce := bytes.Repeat([]byte{0xAA}, 32)
	serverNonce := bytes.Repeat([]byte{0xBB}, 32)

	clientPub, clientPriv, _ := GenerateKeypair()
	serverPub, serverPriv, _ := GenerateKeypair()

	clientSecret, err := SharedSecret(clientPriv, serverPub)
	if err != nil {
		t.Fatalf("client shared secret: %v", err)
	}

	serverSecret, err := SharedSecret(serverPriv, clientPub)
	if err != nil {
		t.Fatalf("server shared secret: %v", err)
	}

	if clientSecret != serverSecret {
		t.Fatal("client and server secret dont match, ECDH failed")
	}

	clientKey, _ := DeriveSessionKey(clientSecret, psk, clientNonce, serverNonce)
	serverKey, _ := DeriveSessionKey(serverSecret, psk, clientNonce, serverNonce)

	if !bytes.Equal(clientKey, serverKey) {
		t.Fatal("session key didnt match handshake will fail")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := bytes.Repeat([]byte{0x22}, 32)
	nonce := bytes.Repeat([]byte{0x01}, 12)

	plaintext := []byte("hehehe supersecret bullshit")
	ciphertext, err := Encrypt(key, nonce, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("what the hell, ciphertext is same as plaintext")
	}

	decrypted, err := Decrypt(key, nonce, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatal("decrypted doesnt match original plaintext")
	}

	badKey := bytes.Repeat([]byte{0x99}, 12)
	_, err = Decrypt(badKey, nonce, ciphertext)
	if err == nil {
		t.Fatal("decrypt with wrong key should fail")
	}

	badNonce := bytes.Repeat([]byte{0x99}, 12)
	_, err = Decrypt(key, badNonce, ciphertext)
	if err == nil {
		t.Fatal("decrypt with wrong nonce should fail")
	}
}

func TestNonceCounter(t *testing.T) {
	nc := NewNonceCounter()

	n1 := nc.Next()
	n2 := nc.Next()
	n3 := nc.Next()

	fmt.Println(n1)
	fmt.Println(n2)
	fmt.Println(n3)

	if len(n1) != 12 {
		t.Fatalf("expected 12 bytes, got %d", len(n1))
	}

	if bytes.Equal(n1, n2) || bytes.Equal(n2, n3) {
		t.Fatal("nonce counter produced duplicate nounces")
	}

	if bytes.Compare(n1, n2) >= 0 {
		t.Fatalf("nonce not increasing")
	}
}
