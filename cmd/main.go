package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
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

	logger.Info("db started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Listen for incoming connections
	listener, err := net.Listen("tcp", cfg.Network.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	logger.Info("server started", "address", cfg.Network.Address)

	// limit the number of connections
	connectionsCount := make(chan bool, cfg.Network.MaxConnections)
	for i := 0; i < cfg.Network.MaxConnections; i++ {
		connectionsCount <- true
	}

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("failed to accept connection", "error", err)
			continue
		}

		// Handle client connection in a goroutine
		go handleClient(ctx, conn, db, logger, connectionsCount)
	}

	// shutdown components
	logger.Warn("database is shutting down")
}

func handleClient(ctx context.Context, conn net.Conn, db *internal.Database, logger *slog.Logger, connectionsCount chan bool) {
	defer func() {
		logger.Warn("main: handle client", "connection close")

		conn.Close()
		connectionsCount <- true // release connection
	}()

	<-connectionsCount

	for {
		requestCtx := context.WithValue(ctx, consts.RequestID, uuid.New().String())
		l := logger.With(consts.RequestID, requestCtx.Value(consts.RequestID))

		connBuf := bufio.NewReader(conn)
		bytes, err := connBuf.ReadBytes('\n')
		if err != nil {
			l.Error("main: read request", "error", err)
			return
		}

		logger.Info("main: incoming request", "data", string(bytes))

		// Process and use the data
		result, err := db.HandleRequest(requestCtx, string(bytes))
		if err != nil {
			logger.Error("main: db: handle request", "error", err)
		}

		resp := fmt.Sprintf("query result: [ %s ] error: [ %v ] \n", result, err)

		_, err = conn.Write([]byte(resp))
		if err != nil {
			logger.Error("main: write response result", "error", err)
			return
		}
	}
}
