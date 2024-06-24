package text

import (
	"context"
	"fmt"
	"net"
)

type Client struct {
	addr string
	conn net.Conn
}

func NewTextClient(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Connect(ctx context.Context) error {
	type result struct {
		conn net.Conn
		err  error
	}
	done := make(chan result, 1)

	if c.conn != nil {
		c.conn.Close()
	}

	go func() {
		conn, err := net.Dial("tcp", c.addr)
		done <- result{conn: conn, err: err}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case r := <-done:
		if r.err != nil {
			return r.err
		}
		c.conn = r.conn
		return nil
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Client) Send(ctx context.Context, data string) (string, error) {
	err := writeWithContext(ctx, c.conn, data)
	if err != nil {
		return "", fmt.Errorf("failed to write data to client: %w", err)
	}

	resp, err := readWithContext(ctx, c.conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response from client: %w", err)
	}

	return resp, nil
}
