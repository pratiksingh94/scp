package main

import (
	"flag"
	"fmt"
	"os"

	// "scp/internal/chat"

	"github.com/pratiksingh94/scp/internal/chat"
	"github.com/pratiksingh94/scp/pkg/scp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: scp <listen|connect> [flags]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "listen":
		listenCmd(os.Args[2:])
	case "connect":
		connectCmd(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func listenCmd(args []string) {
	fs := flag.NewFlagSet("listen", flag.ExitOnError)
	port := fs.String("port", "2008", "port to listen on")
	psk := fs.String("psk", "", "pre-shared key (required)")
	fs.Parse(args)

	if *psk == "" {
		fmt.Fprintln(os.Stderr, "error: --psk is required")
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%s", *port)
	cfg := scp.Config{
		PSK: []byte(*psk),
	}

	fmt.Printf("listening on %s\n", addr)

	listener, err := scp.Listen(addr, &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen error:", err)
		os.Exit(1)
	}

	for {
		fmt.Println("waiting for connection...")
		sess, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "accept error:", err)
			continue
		}

		fmt.Println("client connected, handshake complete yay")
		chat.RunChat(sess)
		fmt.Println("connection close bye")
	}
}

func connectCmd(args []string) {
	fs := flag.NewFlagSet("connecet", flag.ExitOnError)
	host := fs.String("host", "localhost:2008", "host to connect to")
	psk := fs.String("psk", "", "pre-shared key (required)")
	fs.Parse(args)

	if *psk == "" {
		fmt.Fprintln(os.Stderr, "error: --psk is required")
		os.Exit(1)
	}

	cfg := scp.Config{
		PSK: []byte(*psk),
	}
	sess, err := scp.Dial(*host, &cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	fmt.Println("connecting to", host)
	chat.RunChat(sess)
	fmt.Println("connection closed bye")
}
