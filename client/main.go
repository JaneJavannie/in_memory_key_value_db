package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	in := make(chan userInput, 1)

	// want to catch signals without waiting for user input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("ENTER TEXT: ")
		text, err := reader.ReadString('\n')

		in <- userInput{
			text: text,
			err:  err,
		}
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case input := <-in:
			if input.err != nil {
				log.Fatal("input error:", input.err)
			}

			// Send data to the server
			data := []byte(input.text + "\n")
			_, err = conn.Write(data)
			if err != nil {
				log.Fatal("write data to connection:", err)
			}

			// Read and process data from the server
			connBuf := bufio.NewReader(conn)
			bytes, err := connBuf.ReadBytes('\n')
			if err != nil {
				log.Fatal("read bytes from connection:", err)
			}

			fmt.Printf("RESPONSE: %+v\n", string(bytes))
		}
	}

	fmt.Println("bye")
}
