package chat

import (
	"bufio"
	"fmt"
	"os"
	"scp/pkg/scp"
)

func RunChat(sess *scp.Session) {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if err := sess.Send([]byte(line)); err != nil {
				fmt.Fprintln(os.Stderr, "send error: %w", err)
				return
			}
		}
	}()

	for {
		msg, err := sess.Receive()
		if err != nil {
			fmt.Fprintln(os.Stderr, "receive error: %w", err)
			return
		}

		fmt.Printf("< %s\n", msg)
	}
}
