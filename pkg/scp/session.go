package scp

import "fmt"

// Send encrypts plaintext and writes it to the connection as a MsgData packet
// each call uses a unique nonce from an internal counter
func (s *Session) Send(plaintext []byte) error {
	nonce := s.NonceCounter.Next()
	ciphertext, err := Encrypt(s.SessionKey, nonce, plaintext)
	// fmt.Println(string(plaintext))
	// fmt.Println(string(ciphertext))
	if err != nil {
		return fmt.Errorf("sending data: %w", err)
	}

	packet := Packet{
		Type: MsgData,
		Payload: EncodeDataPayload(DataPayload{
			Nonce:      [12]byte(nonce),
			Ciphertext: ciphertext,
		}),
	}

	return WritePacket(s.Conn, packet)
}

// Receive reads the next MsgData packet from the connection and returns the decrypted plaintext
// Blocks until a message arrives or error occurs
func (s *Session) Receive() ([]byte, error) {
	packet, err := ReadPacket(s.Conn)
	if err != nil {
		return nil, err
	}

	if packet.Type == MsgError {
		errPayload, _ := DecodeError(packet.Payload)
		return nil, fmt.Errorf("error from other side: %s", errPayload.Message)
	}

	if packet.Type != MsgData {
		errPacket := Packet{
			Type: MsgError,
			Payload: EncodeError(ErrorPayload{
				Code:    ErrInvalidMessage,
				Message: "expected MsgData, got something else",
			}),
		}

		WritePacket(s.Conn, errPacket)
		return nil, fmt.Errorf("unexpected type: %d", packet.Type)
	}

	dataPayload, err := DecodeDataPayload(packet.Payload)
	if err != nil {
		return nil, err
	}

	return Decrypt(s.SessionKey, dataPayload.Nonce[:], dataPayload.Ciphertext)
}

// Close closes the underlying network connection
func (s *Session) Close() error {
	return s.Conn.Close()
}
