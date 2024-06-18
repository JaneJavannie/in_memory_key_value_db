package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	tcpclient "github.com/JaneJavannie/in_memory_key_value_db/internal/protocol/text"
)

const (
	defaultServerAddress = "127.0.0.1:8088"
)

// --server_address="localhost:8088"

type userInput struct {
	text string
	err  error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Define the command-line options
	address := flag.String("server_address", defaultServerAddress, "db server address")

	// Parse the command-line options
	flag.Parse()

	// Connect to the server
	client := tcpclient.NewTextClient(*address)
	defer client.Close()

	err := HandleInput(ctx, client)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("handle input: " + err.Error())
		}
	}

	slog.Info("bye")
}

func HandleInput(ctx context.Context, client *tcpclient.Client) error {
	err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fmt.Printf("%s: ENTER TEXT: ", time.Now().Format(time.TimeOnly))
		text, err := readRequestStringWithContext(ctx)
		if err != nil {
			return fmt.Errorf("read string: %w", err)
		}

		// Send data to the server

		resp, err := client.Send(ctx, text)
		if err != nil {
			return fmt.Errorf("client send: %w", err)
		}

		fmt.Printf("%s RESPONSE: %+v\n", time.Now().Format(time.TimeOnly), resp)
	}
}

func readRequestStringWithContext(ctx context.Context) (string, error) {
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
