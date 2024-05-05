package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	"in_memory_key_value_db/internal"

	"github.com/google/uuid"
)

const RequestID = "request_id"

func main() {

	db := internal.NewDatabase()

	for i := range 10 {
		_ = i

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")

		text, _ := reader.ReadString('\n')

		c := context.Background()
		ctx := context.WithValue(c, RequestID, uuid.New().String())

		result, err := db.HandleRequest(ctx, text)
		if err != nil {
			slog.Error("db: handle request", RequestID, ctx.Value(RequestID).(string), "error", err)
		}

		println(fmt.Sprintf("%+v", result))
	}
}
