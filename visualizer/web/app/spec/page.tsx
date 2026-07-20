import { Badge } from "@/components/ui/badge"

// prettier my man you deserve hate and love 🔥

function Section({ id, number, title, children }: {
    id: string
    number: string
    title: string
    children: React.ReactNode
}) {
    return (
        <section id={id} className="flex flex-col gap-4">
            <div className="flex items-baseline gap-3">
                <span className="text-muted-foreground text-xs">
                    {number}
                </span>
                <h2 className="text-base font-semibold text-foreground uppercase tracking-wide">{title}</h2>
            </div>

            <div className="flex flex-col gap-3 pl-8 text-sm text-muted-foreground leading-relaxed">
                {children}
            </div>
        </section>
    )
}


function Sub({ number, title, children }: { number: string; title: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-2 mt-2">
            <div className="flex items-baseline gap-2">
                <span className="text-muted-foreground text-xs">{number}</span>
                <h3 className="text-sm font-medium text-foreground">{title}</h3>
            </div>
            <div className="pl-6 flex flex-col gap-2 text-sm text-muted-foreground leading-relaxed">
                {children}
            </div>
        </div>
    )
}

function Code({ children }: { children: React.ReactNode }) {
    return (
        <div className="relative">
            <pre className="bg-muted rounded-md p-4 text-xs text-foreground overflow-x-auto leading-relaxed whitespace-pre scrollbar-thin">
                {children}
            </pre>
        </div>
    )
}

function Table({ headers, rows }: { headers: string[]; rows: string[][] }) {
    return (
        <div className="overflow-x-auto">
            <table className="w-full text-xs border-collapse">
                <thead>
                    <tr className="border-b border-border">
                        {headers.map(h => (
                            <th key={h} className="text-left py-2 pr-6 text-foreground font-medium">{h}</th>
                        ))}
                    </tr>
                </thead>
                <tbody>
                    {rows.map((row, i) => (
                        <tr key={i} className="border-b border-border/40">
                            {row.map((cell, j) => (
                                <td key={j} className="py-2 pr-6 text-muted-foreground">{cell}</td>
                            ))}
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}



export default function SpecPage() {
    return (
        <div className="flex flex-col gap-2 max-w-3xl">

            <div className="flex flex-col gap-4 py-8 border-b border-border mb-8">
                <div className="flex items-center gap-3">
                    <span className="text-xs text-muted-foreground uppercase tracking-widest">Protocol Specification</span>
                </div>

                <h1 className="text-3xl font-bold tracking-light">
                    SCP - Secure Channel Protocol
                </h1>
                <p className="text-sm text-muted-foreground leading-relaxed max-w-xl">
                    A custom TLS-inspired secure channel protocol :3 implementing PSK auth, X25519 ephemeral key exchange and ChaCha20-Poly1305 encryption. Shout out to keeb for name. ^_^
                </p>
            </div>




            {/* TOC  */}
            <div className="border borde-border rounded-md p-4 mb-8">
                <p className="text-xs text-muted-foreground uppercase tracking-widest mb-3">Table of Contents</p>
                <div className="flex flex-col gap-1">
                    {[
                        ["1", "Overview"],
                        ["2", "Packet Format"],
                        ["3", "Message Types"],
                        ["4", "Handshake Protocol"],
                        ["5", "Key Derivation"],
                        ["6", "Data Phase"],
                        ["7", "Cryptographic Primitives"],
                        ["8", "Wire Format Reference"],
                    ].map(([n, t]) => (
                        <a key={n} href={`#section-${n}`} className="text-xs text-muted-foreground hover:text-foreground transition-colors flex gap-3">
                            <span className="w-4">{n}.</span>
                            <span>{t}</span>
                        </a>
                    ))}
                </div>
            </div>




            <div className="flex flex-col gap-12">
                <Section id="section-1" number="1." title="Overview">
                    <p>SCP creates a secure, bidirectional communication channel between two parties over TCP. It provides mutual authentication using a pre-shared key (PSK) and authenticated encryption using ChaCha20-Poly1305.</p>
                    <p>SCP is *NOT* a replacement of TLS, it is an independant protocol i built to learn and demonstrate the core things of a secure channel like handshake, key exchange, key derivation, encryption, etc. This is all implemented in Go.</p>
                    <p>The protocol completes in 2 round trips before the data phase begins. Both parties must have the same PSK before initialising the connection, there is NO certification infrastructure (yet).</p>
                </Section>


                <Section id="section-2" number="2." title="Packet Format">
                    <p>Every SCP message is framed as a packet with 5 byte header followed by a variable length payload.</p>
                    <Code>
                        {`+------------+------------------+----------------------+
| Type (1B)  | Length (4B, BE)  | Payload (N bytes)    |
+------------+------------------+----------------------+`}
                    </Code>

                    <p>
                        <span className="text-foreground">Type</span> - tells the message type (see Section 3)
                        <br />
                        <span className="text-foreground">Length</span> - unsigned 32-bit big-endian integer, length of the payload in bytes
                        <br />
                        <span className="text-foreground">Payload</span> - message specific content
                    </p>
                    <p>
                        The maximum payload size is 2^32 - 1 bytes. You should enforce limits appropriate to your context if you are using the package.
                    </p>
                </Section>



                <Section id="section-3" number="3." title="Message Types">
                    <Table headers={["Value", "Name", "Direction", "Description"]} rows={[
                        ["0x01", "MsgClientHello", "C -> S", "Starts the handshake. Contains client ephemeral public key, none and PSK Proof."],
                        ["0x02", "MsgServerHello", "S -> C", "Server response, contains the same stuff"],
                        ["0x03", "MsgDone", "C <-> S", "Handshake verification, payload is an encrypted MAC proving session key derviation was success"],
                        ["0x04", "MsgData", "C <-> S", "Encrypted application data, contains nonce and ChaCha20-Poly1305 encrypted ciphertext"],
                        ["0x05", "MsgError", "C <-> S", "Protocol error, contains error code and human readable message, terminates connection"]
                    ]} />
                </Section>



                <Section id="section-4" number="4." title="Handshake">
                    <p>The SCP handshake complestes in 2 round trips</p>
                    <Code>{`Client                    Server
  |                         |
  |--- ClientHello -------->|
  |    pub[32] nonce[16]    |
  |    psk_proof[32]        |
  |                         |
  |<-- ServerHello ---------|
  |    pub[32] nonce[16]    |
  |    psk_proof[32]        |
  |                         |
  | [both derive ECDH +     |
  |  session key via HKDF]  |
  |                         |
  |--- Done --------------->|
  |    Enc(HMAC(sk,         |
  |    "client-done"))      |
  |                         |
  |<-- Done ----------------|
  |    Enc(HMAC(sk,         |
  |    "server-done"))      |
  |                         |
  |<=== channel open ======>|`}</Code>

                    <Sub number="4.1" title="PSK Proof - ClientHello">
                        <p>The client computes its PSK proof as:</p>
                        <Code>{`client_psk_proof = HMAC-SHA256(PSK, client_nonce || client_public key)`}</Code>
                        <p>The nonce binds the proof to this specific session, preventing replay attacks.</p>
                    </Sub>

                    <Sub number="4.2" title="PSK Proof - ServerHello">
                        <p>The server computes its PSK proof as:</p>
                        <Code>{`server_psk_proof = HMAC-SHA256(PSK, client_nonce || server_nonce || server_public_key || client_public_key)`}</Code>
                        <p>
                            The server proof includes both nonces and both public keys, binding it to the exact
                            exchange in progress.
                        </p>
                    </Sub>



                    <Sub number="4.3" title="Done Messages">
                        <p>
                            After deriving the session key each party sends a MsgDone packet to prove
                            they derived the correct key. the client sends first:
                        </p>
                        <Code>{`client_done_mac  = HMAC-SHA256(sessionKey, "client-done")
client_done_nonce = nonceCounter.Next()
client_done_payload = ChaCha20-Poly1305_Encrypt(key=sessionKey, nonce=client_done_nonce, plaintext=client_done_mac)`}</Code>
                        <p>The server verifies, then responds symmetrically with <span className="text-foreground">"server-done"</span> as the "domain separator".</p>
                        <p>
                            Domain separation (<span className="text-foreground">"client-done"</span> vs <span className="text-foreground">"server-done"</span>) ensures
                            neither party's Done can be replayed as the other's, even though both use the same session key
                        </p>
                    </Sub>


                    <Sub number="4.4" title="Error Handling">
                        <p>If any verification step fails, the detecing party sends MsgError with the appropriate error code and closes the connection.</p>
                        <Table headers={["Code", "Name", "Condition"]} rows={[
                            ["0x00", "ErrUnknown", "Unspecified error"],
                            ["0x01", "ErrInvalidPSK", "PSK Proof verification failed"],
                            ["0x02", "ErrHandshakeFail", "Done MAC verification failed"],
                            ["0x03", "ErrDecryptFail", "ChaCha20-Poly1305 decryption failed"],
                            ["0x04", "ErrInvalidMessage", "Unexpected message type received"]
                        ]}></Table>
                    </Sub>
                </Section>


                <Section id="section-5" number="5." title="Key Derivation">
                    <p>The session key is derived from ECDH shared secret using HKDF-SHA256:</p>
                    <Code>
                        {`shared_secret  = X25519(my_private_key, peer_public_key)

session_key    = HKDF-SHA256(
  ikm  = shared_secret,          // raw ECDH output
  salt = PSK,                     // authenticates key derivation
  info = "SCP-session"
          || client_nonce
          || server_nonce,        // binds key to this specific session
  len  = 32 bytes
)`}
                    </Code>

                    <p>
                        <span className="text-foreground">ikm</span> - the raw X25519 output is the primary entropy source.
                        HKDF conditions it into a uniform key suitable for ChaCha20-Poly1305
                    </p>

                    <p>
                        <span className="text-foreground">salt = PSK</span> - using the PSK as HKDF salt mixes authentication into key derivation. A session key derived without the correct PSK will differ from the one derived WITH it, even if the ECDH exchange produces same raw output.
                    </p>

                    <p>
                        <span className="text-foreground">info</span> - the context string <span className="text-foreground">"SCP-session"</span> + both nonces make the the derived key unique to to this handshake. Two sessions with the same PSK and same ECDH key but different nonce will produce different session key, so thats never gonna happens.
                    </p>
                </Section>



                <Section id="section-6" number="6." title="Data Phase">
                    <p>
                        After a sucecssful handshake, both parties may exchange MsgData packets in either direction, each packet is independently encrypted and authenticated.
                    </p>

                    <Code>{`to do: add ts`}</Code>
                    <p>
                        <span className="text-foreground">Nonce</span> - a 12-byte big-endian encoded counter, it increments with every sent message and is included in the packet so the receiver can decrypt without maintaning synchronized state.
                    </p>
                    <p>Each party maintains an independent send counter, client and server nonce sequence do not interfere with each other</p>
                </Section>


                <Section id="section-7" number="7." title="Cryptographic Primitives">
                    <Table headers={["Primitive", "Algorithm", "Purpose"]} rows={[
                        ["Key Exchange", "X25519 (RFC 7748)", "Ephemeral ECDH"],
                        ["Symmetric cipher", "ChaCha20-Poly1305 (RFC 8439)", "AEAD encryption"],
                        ["Key Derivation", "HKDF-SHA256 (RFC 5869)", "Session key from ECDH output"],
                        ["MAC", "HMAC-SHA256 (RFC 2104)", "PSK Proofs + Done verification"],
                        ["Handshake nonce", "CSPRNG, 16 bytes", "Session binding"],
                        ["Data nonce", "Counter, 12 bytes", "Encrypting nonce"]
                    ]} />
                </Section>



                <Section id="section-8" number="8." title="Wire Format Reference">
                    <Sub number="9.1" title="MsgClientHello (0x01) - 80 bytes payload">
                        <Code>{`Offset  Size  Field
0       32    client_ephemeral_public_key   X25519 public key
32      16    client_nonce                  Random, 128-bit
48      32    psk_proof                     HMAC-SHA256(PSK, nonce||pub_key)`}</Code>
                    </Sub>
                    <Sub number="9.2" title="MsgServerHello (0x02) - 80 bytes payload">
                        <Code>{`Offset  Size  Field
0       32    server_ephemeral_public_key   X25519 public key
32      16    server_nonce                  Random, 128-bit
48      32    psk_proof                     HMAC-SHA256(PSK, c_nonce||s_nonce||s_pub||c_pub)`}</Code>
                    </Sub>
                    <Sub number="9.3" title="MsgDone (0x03) - 60 bytes payload">
                        <Code>{`Offset  Size  Field
0       12    nonce                         ChaCha20-Poly1305 nonce (counter)
12      48    ciphertext                    Encrypt(sessionKey, nonce, HMAC(sessionKey, role))`}</Code>
                        <p>role is <span className=" text-foreground">"client-done"</span> or <span className=" text-foreground">"server-done"</span>. HMAC output is 32 bytes + 16 byte Poly1305 tag = 48 bytes ciphertext.</p>
                    </Sub>
                    <Sub number="9.4" title="MsgData (0x04) - variable payload">
                        <Code>{`Offset  Size   Field
0       12     nonce        ChaCha20-Poly1305 nonce (send counter)
12      N+16   ciphertext   Encrypt(sessionKey, nonce, plaintext) + 16B auth tag`}</Code>
                    </Sub>
                    <Sub number="9.5" title="MsgError (0x05) - variable payload">
                        <Code>{`Offset  Size  Field
0       1     error_code    See Section 4.4
1       N     message       UTF-8 string, human-readable description`}</Code>
                    </Sub>
                </Section>
            </div>

            <div className="border-t border-border mt-12 pt-6 pb-8">
                <p className="text-muted-foreground">
                    SCP is an educational protocol. Do NOT use in production systems lmao.
                    Implementation: <a href="https://github.com/pratiksingh94/scp" className="text-foreground hover:underline">github.com/pratiksingh94/scp</a>
                </p>
            </div>


        </div>
    )
}