package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
)

func handleInput(ctx context.Context, conn net.Conn) {
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

	for {
		select {
		case <-ctx.Done():
			return
		case input := <-in:
			if input.err != nil {
				log.Fatal("input error:", input.err)
			}

			// Send data to the server
			data := []byte(input.text + "\n")
			if _, err := conn.Write(data); err != nil {
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
}
