# SCP - Secure Channel Protocol


A custom TLS-inspired (i hate TLS) secure channel protocol built in Go.

It establises encrypted and authenticated communication channel using X25519 key exchange, PSK auth and ChaCha20-Poly1305 AEAD encryption.
This isan independent protocol design and NOT a reimplementation of TLS.

**Handshake visualizer:** [scp.pratiksingh.xyz](https://scp.pratiksingh.xyz)    
**Docs (will improve later):** [pkg.go.dev](https://pkg.go.dev/github.com/pratiksingh94/scp/pkg/scp)

> This is an educational project i made for learning and experimenting, do NOT use it in any kind of production environment lmao


## How it works

```
Client                          Server
  |                               |
  |------ ClientHello ----------->|   pub key + nonce + PSK proof
  |                               |
  |<----- ServerHello ------------|   pub key + nonce + PSK proof
  |                               |
  | [both derive session key]     |   X25519 ECDH + HKDF-SHA256
  |                               |
  |------ Done ------------------>|   Enc(HMAC(sessionKey, "client-done"))
  |<----- Done -------------------|   Enc(HMAC(sessionKey, "server-done"))
  |                               |
  |<====== encrypted channel ====>|   ChaCha20-Poly1305
```

Two round trips, both side authenticate before data channel is open.
check out specs page: [/specs](https://scp.pratiksingh.xyz/spec)

---

## Crypto primitives

| Primitive      | Algorithm                    | Purpose                             |
| -------------- | ---------------------------- | ----------------------------------- |
| Key exchange   | X25519 (RFC 7748)            | Ephemeral ECDH                      |
| Encryption     | ChaCha20-Poly1305 (RFC 8439) | AEAD session cipher                 |
| Key derivation | HKDF-SHA256 (RFC 5869)       | Session key from ECDH output        |
| Authentication | HMAC-SHA256 (RFC 2104)       | PSK proofs + handshake verification |


---

## Use as library 

```bash
go get github.com/pratiksingh94/scp
```

```go
import "github.com/pratiksingh94/scp/pkg/scp"
```


**Server:**
```go
listener, err := scp.Listen(":2008", &scp.Config{
    PSK: []byte("pre-shared-secret"),
})

sess, err := listener.Accept()
defer sess.Close()

msg, err := sess.Receive()
fmt.Println(string(msg))

sess.Send([]byte("hewwo from server"))
```

**Client:**
```go
sess, err := scp.Dial("localhost:2008", &scp.Config{
    PSK: []byte("pre-shared-secret"),
})
defer sess.Close()

sess.Send([]byte("haiii from client"))

msg, err := sess.Receive()
fmt.Println(string(msg))
```


---

## CLI

Yeah so i made this small CLI chat app on top of the protocol

Build:
```bash
go build -o scp ./cmd/scp
```

Listen:
```bash
./scp listen --port 2008 --psk mysecret
```

Connect:
```bash
./scp connect --host hostname:2008 --psk mysecret
```

Both sides read from stdin and print received messages, and it works over a real network too, tested between a local machine and a remote VPS.


---

## Some Security Properties

- **Mutual authentication** - both sides prove PSK knowledge before the channel opens
- **Forward secrecy** - ephemeral X25519 keys are "discarded"  after each handshake (generated fresh for each handshake)
- **Replay protection** - client and server nonces bind PSK proofs to a specific session
- **Authenticated encryption** - ChaCha20-Poly1305 detects any tampering with ciphertext

SCP does NOT provide: certificate infrastructure (yet), post-quantum security, session resumption,
or identity beyond PSK possession.
Check out specs page on visualizer for more details on everything

---

Built for [stardance](https://stardance.hackclub.com/projects/153)!