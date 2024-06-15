package configs

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"gopkg.in/yaml.v3"
)

type Logger struct {
	Level    string `yaml:"level"`
	IsPretty bool   `yaml:"is_pretty"`
}

type Wal struct {
	FlushingBatchSize    int           `yaml:"flushing_batch_size"`
	FlushingBatchTimeout time.Duration `yaml:"flushing_batch_timeout"`
	MaxSegmentSize       string        `yaml:"max_segment_size"`
	MaxSegmentSizeBytes  int           `yaml:"max_segment_size_bytes"`
	DataDir              string        `yaml:"data_directory"`
}

type Network struct {
	Address        string `yaml:"address"`
	MaxConnections int    `yaml:"max_connections"`
}

type Engine struct {
	Type string `yaml:"type"`
}

type App struct {
	Timeout time.Duration `yaml:"timeout"`
}

type Config struct {
	App     App     `yaml:"app"`
	Engine  Engine  `yaml:"engine"`
	Network Network `yaml:"network"`
	Wal     *Wal    `yaml:"wal"`
	Logger  Logger  `yaml:"logger"`
}

func NewConfig(configPath string) (*Config, error) {
	if err := validateConfigPath(configPath); err != nil {
		return nil, fmt.Errorf("validate config path: %w", err)
	}

	cfg, err := parseConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}

	err = cfg.SetDefaults()
	if err != nil {
		return nil, fmt.Errorf("set default config: %w", err)
	}

	return cfg, nil
}

func (c *Config) SetDefaults() error {
	if c.App.Timeout == 0 {
		c.App.Timeout = consts.AppTimeout * time.Second
	}
	if c.Engine.Type == "" {
		c.Engine.Type = consts.EngineType
	}
	if c.Network.Address == "" {
		c.Network.Address = consts.ServerAddress
	}
	if c.Network.MaxConnections == 0 {
		c.Network.MaxConnections = consts.MaxConnections
	}
	if c.Logger.Level == "" {
		c.Logger.Level = consts.LogLevel
	}

	if c.Wal != nil {
		if c.Wal.FlushingBatchSize == 0 {
			c.Wal.FlushingBatchSize = consts.WalFlushingBatchSize
		}
		if c.Wal.FlushingBatchTimeout == 0 {
			c.Wal.FlushingBatchTimeout = consts.WalFlushingBatchTimeout
		}
		if c.Wal.MaxSegmentSize == "" {
			c.Wal.MaxSegmentSize = consts.WalMaxSegmentSize
		}
		if c.Wal.DataDir == "" {
			c.Wal.DataDir = consts.WalDataDir
		}

		bytesSize, err := parseToBytes(c.Wal.MaxSegmentSize)
		if err != nil {
			return fmt.Errorf("parse wal max segment size to bytes: %w", err)
		}
		c.Wal.MaxSegmentSizeBytes = bytesSize
	}

	return nil
}

func parseToBytes(sizeString string) (int, error) {
	count := ""
	measurment := ""

	for _, sym := range sizeString {
		if isDigit(sym) {
			count = count + string(sym)
		}
		if isLetter(sym) {
			measurment = measurment + string(sym)
		}
	}

	measurment = strings.ToUpper(measurment)

	size, err := strconv.Atoi(count)
	if err != nil {
		return 0, fmt.Errorf("parse size: %w", err)
	}

	switch measurment {
	case "B":
		slog.Info("no need to convert bytes")
	case "KB":
		size = size * 1024
	case "MB":
		size = size * 1024 * 1024
	case "GB":
		size = size * 1024 * 1024 * 1024
	case "TB":
		size = size * 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("measurement type not supported: %s", measurment)
	}

	return size, nil
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func parseConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode config file: %w", err)
	}

	return config, nil
}

func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
