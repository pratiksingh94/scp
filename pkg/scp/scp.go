// Package spc implements the Secure Channel Protocol, a custom TLS-inspired, check the github for more info.
package scp

import (
	"fmt"
	"net"
)

// Config holds the configuration for an SCP connection, well duh.
type Config struct {
	// PSK is the pre-shared key used for mutual authentication.
	// Both sides must use the same PSK.
	PSK []byte
}

// Dial connects to an SCP server at addr and performs the handshake
// Returns an established Session ready for sending and receiving data
// Handshake verifies mutual PSK knowledge before returning
func Dial(addr string, cfg *Config) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return ClientHandshake(conn, cfg.PSK)
}

// Listener accepts incoming SCP connections on a TCP address
type Listener struct {
	inner net.Listener
	cfg   *Config
}

// Listen creates a Listener on the given TCP address
// Each accepted connection will perform the SCP server handshake
// using the PSK in cfg before a Session is returned
func Listen(addr string, cfg *Config) (*Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listening: %w", err)
	}

	return &Listener{
		inner: l,
		cfg:   cfg,
	}, nil
}

// Accept waits for an incoming connection and perform the handshake
// returning an established Session or an error if the handshake fail
func (l *Listener) Accept() (*Session, error) {
	conn, err := l.inner.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept: %w", err)
	}

	return ServerHandshake(conn, l.cfg.PSK)
}

// Close stops the Listener from accepting new connection
func (l *Listener) Close() error {
	return l.inner.Close()
}
