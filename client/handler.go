package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal"
)

func handleInput(ctx context.Context, conn net.Conn) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fmt.Printf("ENTER TEXT: ")
		text, err := readStringWithContext(ctx)
		if err != nil {
			return fmt.Errorf("read string: %w", err)
		}

		println(time.Now().Format(time.TimeOnly))

		// Send data to the server
		data := []byte(text + "\n")
		if _, err := conn.Write(data); err != nil {
			return fmt.Errorf("write string: %w", err)
		}

		// Read and process data from the server
		bytes, err := internal.ReadBytesWithContext(ctx, conn)
		if err != nil {
			return fmt.Errorf("read bytes: %w", err)
		}

		fmt.Printf("%s RESPONSE: %+v\n", time.Now().Format(time.TimeOnly), string(bytes))
	}
}

func readStringWithContext(ctx context.Context) (string, error) {
	type result struct {
		text string
		err  error
	}
	done := make(chan result, 1)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		done <- result{
			text: text,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-done:
		return r.text, r.err
	}
}
