package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/JaneJavannie/in_memory_key_value_db/internal"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	mylogger "github.com/JaneJavannie/in_memory_key_value_db/internal/logger"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/storage/engine"
	wals "github.com/JaneJavannie/in_memory_key_value_db/internal/wal"
)

const (
	defaultMasterConfigPath = "./config.yaml"
	defaultSlaveConfigPath  = "./config_slave.yaml"
)

// --config=./config.yaml
// --config=./config_slave.yaml

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Define the command-line options
	configPath := flag.String("config", defaultSlaveConfigPath, "config path")

	// Parse the command-line options
	flag.Parse()

	cfg, err := configs.NewConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := mylogger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("config loaded")

	storage, err := engine.NewInMemoryStorage(cfg.Wal)
	if err != nil {
		log.Fatal(err)
	}

	replicationType := cfg.Replication.Type

	wal, err := wals.NewWal(logger, cfg.Wal, replicationType)
	if err != nil {
		log.Fatal(err)
	}
	wal.Start(cfg.Wal)

	inMemoryEngine, err := engine.NewInMemoryEngine(storage, wal, logger, cfg.Wal, replicationType)
	if err != nil {
		log.Fatal(err)
	}

	db, err := internal.NewDatabase(inMemoryEngine, logger)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("db configured")

	server := internal.NewTcpServer(cfg.Network.MaxConnections, cfg.Network.Address, db, logger)
	logger.Info("server configured")

	err = server.Start()
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

	wal.Stop(cfg.Wal)

	logger.Warn("bb")
}
