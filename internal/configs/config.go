package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Logger struct {
	Level    string `yaml:"level"`
	IsPretty bool   `yaml:"is_pretty"`
}

type Config struct {
	App struct {
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"app"`
	Logger Logger `yaml:"logger"`
}

func NewConfig(configPath string) (*Config, error) {
	if err := validateConfigPath(configPath); err != nil {
		return nil, fmt.Errorf("validate config path: %w", err)
	}

	cfg, err := parseConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}

	return cfg, nil
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
