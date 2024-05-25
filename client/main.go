package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
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
	conn, err := net.Dial("tcp", *address)
	if err != nil {
		log.Fatal("dial tcp:", err)
	}
	defer conn.Close()

	handleInput(ctx, conn)

	fmt.Println("bye")
}
