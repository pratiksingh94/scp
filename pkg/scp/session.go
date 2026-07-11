package scp

import "fmt"

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

func (s *Session) Close() error {
	return s.Conn.Close()
}
