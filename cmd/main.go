package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"

	"in_memory_key_value_db/internal"
	"in_memory_key_value_db/internal/consts"

	"github.com/google/uuid"
)

func main() {

	db := internal.NewDatabase()

	for i := range 10 {
		_ = i

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")

		text, _ := reader.ReadString('\n')

		c := context.Background()
		ctx := context.WithValue(c, consts.RequestID, uuid.New().String())

		slog.Info("main: incoming request", consts.RequestID, ctx.Value(consts.RequestID).(string))

		result, err := db.HandleRequest(ctx, text)
		if err != nil {
			slog.Error("db: handle request", consts.RequestID, ctx.Value(consts.RequestID).(string), "error", err)
		}

		println(fmt.Sprintf("%+v", result))
	}
}
