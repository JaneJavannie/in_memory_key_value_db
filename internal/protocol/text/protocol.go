package text

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
)

func writeWithContext(ctx context.Context, conn net.Conn, data string) error {
	type result struct {
		err error
	}
	done := make(chan result, 1)

	go func() {
		_, err := conn.Write([]byte(data + "\r"))

		if err != nil {
			done <- result{
				err: fmt.Errorf("failed to write data to client: %w", err),
			}

			return
		}

		done <- result{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case r := <-done:
		return r.err
	}
}

func readWithContext(ctx context.Context, conn net.Conn) (string, error) {
	type result struct {
		bytes []byte
		err   error
	}
	done := make(chan result, 1)

	go func() {
		connBuf := bufio.NewReader(conn)
		bytes, err := connBuf.ReadBytes('\r')
		done <- result{
			bytes: bytes,
			err:   err,
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-done:
		resp := strings.TrimSuffix(string(r.bytes), "\r")
		return resp, r.err
	}
}
