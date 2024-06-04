package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/JaneJavannie/in_memory_key_value_db/internal"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	mylogger "github.com/JaneJavannie/in_memory_key_value_db/internal/logger"
)

const configPath = "./config.yaml"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := configs.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := mylogger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config loaded")

	db := internal.NewDatabase(logger)
	logger.Info("db configured")

	server := internal.NewTcpServer(cfg.Network.MaxConnections, cfg.Network.Address, db, logger)
	logger.Info("server configured")

	err = server.Start(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	<-ctx.Done()

	logger.Info("database is shutting down...")

	// shutdown components
	err = server.Stop()
	if err != nil {
		logger.Warn("server stop: %v", err)
	}

	logger.Warn("bb")
}
