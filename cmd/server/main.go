package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/JaneJavannie/in_memory_key_value_db/internal"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/configs"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	mylogger "github.com/JaneJavannie/in_memory_key_value_db/internal/logger"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/protocol/text"
	"github.com/JaneJavannie/in_memory_key_value_db/internal/storage/engine"
	wals "github.com/JaneJavannie/in_memory_key_value_db/internal/wal"
	"github.com/JaneJavannie/in_memory_key_value_db/replication"
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

	storage, err := engine.NewInMemoryStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	replicationCfg := cfg.Replication

	replicationType := ""
	masterAddress := ""
	if replicationCfg != nil {
		masterAddress = replicationCfg.MasterAddress
	}

	client := text.NewTextClient(masterAddress)
	replicationServer := text.NewTcpServer(consts.ReplicationMaxConnections, masterAddress, logger)

	if replicationCfg != nil {
		replicationType = replicationCfg.Type
		masterAddress = replicationCfg.MasterAddress

		newReplication, err := replication.NewReplication(cfg, client, replicationServer, storage, logger)
		if err != nil {
			log.Fatal(err)
		}

		err = newReplication.Start(ctx, replicationCfg.SyncInterval)
		if err != nil {
			log.Fatal(err)
		}

		logger.Info("replication started")
	}

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

	server := text.NewTcpServer(cfg.Network.MaxConnections, cfg.Network.Address, logger)
	server.SetOnReceive(func(ctx context.Context, request string) string {
		// Process the data
		result, err := db.HandleRequest(ctx, request)
		if err != nil {
			logger.Error("handle client: db: handle request", "error", err)
		}
		resp := fmt.Sprintf("query result: [ %s ] error: [ %v ] \n", result, err)

		return resp
	})

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

	err = replicationServer.Stop()
	if err != nil {
		logger.Warn("server stop: %v", err)
	}

	wal.Stop(cfg.Wal)

	logger.Warn("bb")
}
