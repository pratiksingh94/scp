# SPEC

just deciding on things for now

## Suite
just one set of stuff, dont wanna make it too complicated

```
Key exchange: X25519 ECDH
Encryption: ChaCha20-Poly1305
Authentication: Pre-shared key (i MIGHT add certs later)
Handshake: 2-RTT design
```

i am gonna use ChaCha20-Poly1305 instead of AES-GCM just because chacha sounds more funny, i know it AES can be faster on systems which have AES optimization instructions built in and how chacha is more modern, but lowkey i dont care about any of that.


## Packet format
so i am gonna keep it simple for now
it will be using binary cuz its real shit

```
[1 byte: msg type] [4 bytes: payload length] [N bytes: payload]
```

### Message Types
```
0x01 = ClientHello
0x02 = ServerHello
0x03 = Done
0x04 = Data
0x05 = Error
```


---

## Handshake
i will be making it like TLS itself, but not that complicated at least for now
it will be a 2-RTT handshake

and uhhh yes ofc it will maintain a forward secrecy to prevent "store now, decrypt later" attacks so yes it will use ephemeral keys

```
Client                    Server
  |                         |
  |---- ClientHello ------->|  (client ephemeral public key + PSK proof)
  |                         |
  |<--- ServerHello --------|  (server ephemeral public key + PSK proof)
  |                         |
  |  [both derive session   |
  |   keys from ECDH +PSK]  |
  |                         |
  |--------- Done --------->|  (encrypted with session key, proves client has key)
  |                         |
  |--------- Done --------->|  (same from server)
  |                         |
  |<======= Data ==========>|  (encrypted channel open)
```