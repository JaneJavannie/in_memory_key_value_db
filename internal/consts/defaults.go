package consts

import "time"

const (
	AppTimeout          = 3
	EngineType          = "in_memory"
	MasterServerAddress = "127.0.0.1:8088"
	MaxConnections      = 10

	ReplicationSyncInterval = 5 * time.Second
	ReplicationTypeSlave    = "slave"
	SlaveServerAddress      = "127.0.0.1:8089"

	LogLevel       = "info"
	MaxMessageSize = 1024

	WalMaxSegmentSize       = "10MB"
	WalFlushingBatchSize    = 100
	WalFlushingBatchTimeout = 10 * time.Millisecond
	WalDataDir              = "/data/wal"
)
