package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JaneJavannie/in_memory_key_value_db/internal"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	mylogger "github.com/JaneJavannie/in_memory_key_value_db/internal/logger"

	"github.com/google/uuid"
)

const configPath = "./config.yaml"

func main() {
	cfg, err := configs.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := mylogger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatal(err)
	}

	db := internal.NewDatabase(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = listenUserInput(ctx, logger, db)
	if err != nil {
		log.Fatal(err)
	}

	// shutdown components
	logger.Warn("database is shutting down")
}

func listenUserInput(ctx context.Context, logger *slog.Logger, db *internal.Database) error {
	type userInput struct {
		text string
		err  error
	}

	for {
		in := make(chan userInput, 1)

		go func() {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("ENTER TEXT: ")

			text, err := reader.ReadString('\n')
			input := userInput{
				text: text,
				err:  err,
			}

			in <- input
		}()

		select {
		case <-ctx.Done():
			return nil
		case input := <-in:
			if input.err != nil {
				return input.err
			}

			requestCtx := context.WithValue(ctx, consts.RequestID, uuid.New().String())

			logger.Info("main: incoming request", consts.RequestID, requestCtx.Value(consts.RequestID).(string))

			result, err := db.HandleRequest(requestCtx, input.text)
			if err != nil {
				logger.Error("db: handle request", consts.RequestID, requestCtx.Value(consts.RequestID).(string), "error", err)
			}

			fmt.Printf("RESPONSE: %+v\n", result)
		}
	}
}
