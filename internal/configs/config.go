package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
	"gopkg.in/yaml.v3"
)

type Logger struct {
	Level    string `yaml:"level"`
	IsPretty bool   `yaml:"is_pretty"`
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

	cfg.SetDefaults()

	return cfg, nil
}

func (c *Config) SetDefaults() {
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
