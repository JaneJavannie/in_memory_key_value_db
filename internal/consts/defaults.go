package consts

import "time"

const (
	AppTimeout     = 3
	EngineType     = "in_memory"
	ServerAddress  = "localhost:8088"
	MaxConnections = 10
	LogLevel       = "info"
	MaxMessageSize = 1024

	WalMaxSegmentSize       = "10MB"
	WalFlushingBatchSize    = 100
	WalFlushingBatchTimeout = 10 * time.Millisecond
	WalDataDir              = "/data/wal"
)
