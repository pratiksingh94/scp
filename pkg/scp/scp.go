package scp

import (
	"fmt"
	"net"
)

type Config struct {
	PSK []byte
}

func Dial(addr string, cfg *Config) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return ClientHandshake(conn, cfg.PSK)
}

type Listener struct {
	inner net.Listener
	cfg   *Config
}

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

func (l *Listener) Accept() (*Session, error) {
	conn, err := l.inner.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept: %w", err)
	}

	return ServerHandshake(conn, l.cfg.PSK)
}

func (l *Listener) Close() error {
	return l.inner.Close()
}
