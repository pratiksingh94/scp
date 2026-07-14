package main

// yes this is all fake, just a simulation fr no packets are being sent

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"scp/pkg/scp"
	"time"
)

type Actor string

const (
	ActorClient Actor = "client"
	ActorServer Actor = "server"
	ActorBoth   Actor = "both"
)

type VisualizerStep struct {
	Step       int            `json:"step"`
	Actor      Actor          `json:"actor"`
	Type       string         `json:"type"`
	Title      string         `json:"string"`
	Data       map[string]any `json:"data"`
	Annotation string         `json:"annotation"`
	IsTransmit bool           `json:"is_transmit"`
	Phase      string         `json:"phase"`
}

var demoPSK = []byte("demo-psk")

func RunSimulation(emit func(VisualizerStep)) {
	step := 0

	next := func(actor Actor, typ, title, phase string, data map[string]any, annotation string, isTransmit bool) {
		step++
		emit(VisualizerStep{
			Step:       step,
			Actor:      actor,
			Type:       typ,
			Title:      title,
			Data:       data,
			Annotation: annotation,
			IsTransmit: isTransmit,
			Phase:      phase,
		})

		time.Sleep(800 * time.Millisecond)
	}

	clientPub, clientPriv, _ := scp.GenerateKeypair()
	next(ActorClient, "keypair_generated", "Client generates X25519 keypair", "handshake", map[string]any{
		"public_key":  hex.EncodeToString(clientPub[:]),
		"private_key": "(kept secret, never transmit it)",
	}, "X25519 is an elliptic cruve Diffie-Hellman function. Each session generates a fresh keypair, even if old session key is stolen, past sessions stay secure. This property is called forward secrecy. :3", false)

	clientNonce, _ := scp.RandomNonce(16)
	next(ActorClient, "nonce_generated", "Client generates random nonce", "handshake", map[string]any{
		"nonce": hex.EncodeToString(clientNonce),
		"size":  "16 bytes (128 bits of entropy)",
	}, "The nonce \"binds\" this PSK proof to a specific session. Without it an attacker could record a valid ClientHello and replay it later to impersonate a client, this is called a replay attack.", false)

	clientPSKProof := scp.ComputePSKProof(demoPSK, clientNonce, clientPub[:])
	next(ActorClient, "psk_proof_computed", "Client computes PSK Proof", "handshake", map[string]any{
		"formula": "HMAC-SHA256(PSK, clientNonce + clientPublicKey)",
		"result":  hex.EncodeToString(clientPSKProof),
	}, "This PSK proves the client knows the pre-shared secret without revealing it. HMAC is a one-way function so you can verify it with same key but cant reverse it to get the key itself.", false)

	next(ActorClient, "client_hello_sent", "Client -> Server: ClientHello", "handshake", map[string]any{
		"packet_type":  "0x01 (MsgClientHello)",
		"payload_size": "80 bytes",
		"structure":    "public_key[32] + psk_proof[32] + nonce[16]",
		"public_key":   hex.EncodeToString(clientPub[:]),
		"psk_proof":    hex.EncodeToString(clientPSKProof),
		"nonce":        hex.EncodeToString(clientNonce),
	}, "The ClientHello is the first message of the handshake, it contains everything that server needs to verify client's identity.", true)

	next(ActorServer, "client_hello_received", "Server receives ClientHello", "handshake", map[string]any{
		"parsed_public_key": hex.EncodeToString(clientPub[:]),
		"parsed_psk_proof":  hex.EncodeToString(clientPSKProof),
		"parsed_nonce":      hex.EncodeToString(clientNonce),
	}, "The server parses the incoming hello packet.", false)

	expectedClientProof := scp.ComputePSKProof(demoPSK, clientNonce, clientPub[:])
	proofValid := bytes.Equal(expectedClientProof, clientPSKProof)
	next(ActorServer, "psk_verified", "Server verifies client PSK proof", "handshake", map[string]any{
		"recomputed": hex.EncodeToString(expectedClientProof),
		"received":   hex.EncodeToString(clientPSKProof),
		"match":      fmt.Sprintf("%v ✓", proofValid),
	}, "The server idependently recomputes the expected proof using its own copy of PSK, if it matches, the client knows the secret. If not, an error is sent and connection is terminated.", false)

	serverPub, serverPriv, _ := scp.GenerateKeypair()
	next(ActorServer, "keypair_generated", "Server generates ephemeral X25519 keypair", "handshake", map[string]any{
		"public_key":  hex.EncodeToString(serverPub[:]),
		"private_key": "(kept secret, shhhh)",
	}, "The server also generates a fresh keypair for this session.", false)

	serverNonce, _ := scp.RandomNonce(16)
	next(ActorServer, "nonce_generated", "Server generates random nonce", "handshake", map[string]any{
		"nonce": hex.EncodeToString(serverNonce),
		"size":  "16 bytes (128 bits of entropy)",
	}, "The same role as client nonce, ensuring session is unique even if one side's nonce gets repeated between sessions somehow", false)

	serverPSKProof := scp.ComputePSKProof(demoPSK, clientNonce, serverNonce, serverPub[:], clientPub[:])
	next(ActorServer, "psk_proof_computed", "Server computes PSK Proof", "handshake", map[string]any{
		"formula": "HMAC-SHA256(PSK, clientNonce + serverNonce + clientPublicKey + serverPublicKey)",
		"result":  hex.EncodeToString(serverPSKProof),
	}, "The server's proof includes both nonces and both public key, this makes it bound to THIS exchange only.", false)

	next(ActorServer, "server_hello_sent", "Server -> Client: ServerHello", "handshake", map[string]any{
		"packet_type":  "0x02 (MsgServerHello)",
		"payload_size": "80 bytes",
		"structure":    "public_key[32] + psk_proof[32] + nonce[16]",
		"public_key":   hex.EncodeToString(serverPub[:]),
		"psk_proof":    hex.EncodeToString(serverPSKProof),
		"nonce":        hex.EncodeToString(serverNonce),
	}, "The ServerHello has the same structure as ClientHello, now client can also verify server's identity and compute the shared secret.", true)

	next(ActorClient, "server_hello_received", "Client receives ServerHello", "handshake", map[string]any{
		"parsed_public_key": hex.EncodeToString(serverPub[:]),
		"parsed_psk_proof":  hex.EncodeToString(serverPSKProof),
		"parsed_nonce":      hex.EncodeToString(serverNonce),
	}, "The client now has the server's emphemeral public key and nonce, thats the final input needed to complete the key exchange.", false)

	expectedServerProof := scp.ComputePSKProof(demoPSK, clientNonce, serverNonce, serverPub[:], clientPub[:])
	serverProofValid := bytes.Equal(expectedServerProof, serverPSKProof)
	next(ActorClient, "psk_verified", "Client verifies server PSK proof", "handshake", map[string]any{
		"recomputed": hex.EncodeToString(expectedServerProof),
		"received":   hex.EncodeToString(serverPSKProof),
		"match":      fmt.Sprintf("%v ✓", serverProofValid),
	}, "The client also verifies the server, neither side trusts each other until both prove their PSK.", false)

	clientSecret, _ := scp.SharedSecret(clientPriv, serverPub)
	serverSecret, _ := scp.SharedSecret(serverPriv, clientPub)
	secretMatch := clientSecret == serverSecret
	next(ActorBoth, "ecdh_computed", "Both sides independetly compute ECDH shared secret", "handshake", map[string]any{
		"client_computes": "X25519(clientPrivateKey, serverPublicKey)",
		"server_computes": "X25519(serverPrivateKey, clientPublicKey)",
		"shared_secret":   hex.EncodeToString(clientSecret[:]),
		"secret_matched":  fmt.Sprintf("%v ✓", secretMatch),
		"transmitted":     "never, it is derived independently on each side",
	}, "This is basically the core of Diffie-Hellman: two parties arrive at the same secret by exchanging only public information, an eavesdropper who saw the entire conversation still cannot compute the shared secret", false)

	clientSessionKey, _ := scp.DeriveSessionKey(clientSecret, demoPSK, clientNonce, serverNonce)
	serverSessionKey, _ := scp.DeriveSessionKey(serverSecret, demoPSK, clientNonce, serverNonce)
	keysMatch := bytes.Equal(clientSessionKey, serverSessionKey)
	next(ActorBoth, "session_key_derived", "Both side derive session key using HKDF", "handshake", map[string]any{
		"formula":     "HKDF-SHA256(ikm=sharedSecret, salt=PSK, info='SCP-session'+clientNonce+serverNonce)",
		"session_key": hex.EncodeToString(clientSessionKey),
		"keys_match":  fmt.Sprintf("%v ✓", keysMatch),
	}, "HKDF (HMAC-Based Key Derivation Function) takes the raw ECDH output and produces a nice session key. Mixing the nonces binds the session key to this session", false)

	clientDoneMAC := scp.ComputePSKProof(clientSessionKey, []byte("client-done"))
	next(ActorClient, "done_mac_computed", "Client computes handshake verification MAC", "handshake", map[string]any{
		"formula": "HMAC-SHA256(sessionKey, 'client-done')",
		"mac":     hex.EncodeToString(clientDoneMAC),
	}, "'client-done' is a domain seperator type shi, it ensures this MAC can never be confused with server's Done MAC even though both use same session key.", false)

	clientNC := scp.NewNonceCounter()
	clientDoneNonce := clientNC.Next()
	encryptedClientDone, _ := scp.Encrypt(clientSessionKey, clientDoneNonce, clientDoneMAC)
	next(ActorClient, "done_encrypted", "Client encryptes Done with session key", "handshake", map[string]any{
		"algorithm":  "ChaCha20-Poly1305",
		"plaintext":  hex.EncodeToString(clientDoneMAC),
		"ciphertext": hex.EncodeToString(encryptedClientDone),
		"nonce":      hex.EncodeToString(clientDoneNonce),
	}, "This is the first use of session key, we are using ChaCha20-Poly1305 which is an authenticated encryption algorithm (AEAD), cuz it simulatenously encrypts and authenticates", false)

	next(ActorClient, "done_sent", "Client -> Server: Done", "handshake", map[string]any{
		"packet_type": "0x03 (MsgDone)",
		"structure":   "nonce[12] + ciphertext[48]",
		"nonce":       hex.EncodeToString(clientDoneNonce),
		"ciphertext":  hex.EncodeToString(encryptedClientDone),
	}, "The Done packet proves the client derived the correct session key. The nonce is sent in plaintext so the server can decrypt the packet (but without session key the ciphertext doesnt reveal anything).", true)

	serverNC := scp.NewNonceCounter()
	decryptedClientDone, _ := scp.Decrypt(serverSessionKey, serverNC.Next(), encryptedClientDone)
	expectedClientDone := scp.ComputePSKProof(serverSessionKey, []byte("client-done"))
	clientDoneValid := bytes.Equal(decryptedClientDone, expectedClientDone)
	next(ActorServer, "done_verified", "Server decrypts and verifies Client Done", "handshake", map[string]any{
		"decrypted": hex.EncodeToString(decryptedClientDone),
		"expected":  hex.EncodeToString(expectedClientDone),
		"verified":  fmt.Sprintf("%v ✓", clientDoneValid),
	}, "If the decryption succeeds and the MAC matches, it proves the server knows: (1) the client derived the same session key, (2) the handshake was not tampered, (3) both sides are not ready for data phase.", false)

	serverDoneMAC := scp.ComputePSKProof(serverSessionKey, []byte("server-done"))
	serverDoneNonce := serverNC.Next()
	encryptedServerDone, _ := scp.Encrypt(serverSessionKey, serverDoneNonce, serverDoneMAC)
	next(ActorServer, "done_encrypted", "Server computes and encryptes Done", "handshake", map[string]any{
		"formula":    "HMAC-SHA256(sessionKey, 'server-done')",
		"mac":        hex.EncodeToString(serverDoneMAC),
		"nonce":      hex.EncodeToString(serverDoneNonce),
		"ciphertext": hex.EncodeToString(encryptedServerDone),
	}, "The server's Done uses 'server-done' as a seperator. This prevents the server's done to be replayed.", false)

	// start from step 20
	next(ActorServer, "done_sent", "Server -> Client: Done", "handshake", map[string]any{
		"packet_type": "0x03 (MsgDone)",
		"structure":   "nonce[12] + ciphertext[48]",
		"nonce":       hex.EncodeToString(serverDoneNonce),
		"ciphertext":  hex.EncodeToString(encryptedServerDone),
	}, "Server's Done completes mutual verification. After sending this server side's handshake is considered complete.", true)

	decryptedServerDone, _ := scp.Decrypt(clientSessionKey, clientNC.Next(), encryptedServerDone)
	expectedServerDone := scp.ComputePSKProof(clientSessionKey, []byte("server-done"))
	serverDoneValid := bytes.Equal(decryptedServerDone, expectedServerDone)
	next(ActorClient, "done_verified", "Client decrypts and verifies server Done", "handshake",
		map[string]any{
			"decrypted": hex.EncodeToString(decryptedServerDone),
			"expected":  hex.EncodeToString(expectedServerDone),
			"verified":  fmt.Sprintf("%v ✓", serverDoneValid),
		},
		"Mutual verification is complete. Both sides confirmed they derived same session keys. The handshake is finished yay :3",
		false,
	)

	next(ActorBoth, "handshake_complete", "Handshake complete!", "handshake", map[string]any{
		"session_key":    hex.EncodeToString(clientSessionKey),
		"cipher":         "ChaCha20-Poly1305",
		"key_exchange":   "X25519 ECDH",
		"authentication": "PSK + HMAC-SHA256",
		"key_derivation": "HKDF-SHA256",
		"round_trips":    "2 RTT",
	}, "The ephemeral private keys are not used again, even if the session key is later compromised, the keys needed to rederive it are gone, this is forward secrecy. ", false)

	time.Sleep(1000 * time.Millisecond)

	// data phase

	clientDataNC := scp.NewNonceCounter()
	serverDataNC := scp.NewNonceCounter()

	plaintext := []byte("hai :3 this is client!")
	msgNonce := clientDataNC.Next()
	encryptedMsg, _ := scp.Encrypt(clientSessionKey, msgNonce, plaintext)

	next(ActorClient, "data_sent", "Client -> Server: Encrypted message", "data", map[string]any{
		"plaintext":  string(plaintext),
		"nonce":      hex.EncodeToString(msgNonce),
		"ciphertext": hex.EncodeToString(encryptedMsg),
	}, "Each data packet includes a nonce and the ciphertext. The nonce is like a counter, it increments every message.", true)

	decryptedMsg, _ := scp.Decrypt(serverSessionKey, serverDataNC.Next(), encryptedMsg)
	next(ActorServer, "data_received", "Server decrypts message", "data", map[string]any{
		"ciphertext": hex.EncodeToString(encryptedMsg),
		"plaintext":  string(decryptedMsg),
	}, "Decryption also verifies a 16-byte Poly1305 authentication tag which makes sure the ciphertext is not modified in transit.", false)

	reply := []byte("secure channel is so tuff uwu")
	replyNonce := serverDataNC.Next()
	encryptedReply, _ := scp.Encrypt(serverSessionKey, replyNonce, reply)
	next(ActorServer, "data_sent", "Server -> Client: Encrypted reply", "data", map[string]any{
		"plaintext":  string(reply),
		"nonce":      hex.EncodeToString(replyNonce),
		"ciphertext": hex.EncodeToString(encryptedReply),
	}, "The server uses its own idependent nonce counter. These are maintained seperately and not synced which can cause probems in bidirectional comms.", false)

	decReply, _ := scp.Decrypt(clientSessionKey, clientDataNC.Next(), encryptedReply)
	next(ActorClient, "data_received", "Client decrypts reply", "data", map[string]any{
		"ciphertext": hex.EncodeToString(encryptedReply),
		"plaintext":  string(decReply),
	}, "The full exchange is complete. Both sides communicated securely, the PSK and session key were never transmitted, all data was encrypted and authenticated, and forward secrecy ensures past sessions are safe even if future keys are compromised. TUFF SHI 🔥.", false)

}
