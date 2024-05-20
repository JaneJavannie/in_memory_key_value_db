package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/JaneJavannie/in_memory_key_value_db/internal/consts"
)

func TestNewConfig(t *testing.T) {
	type test struct {
		input string
		want  *Config
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get current working dir: %v", err)
	}

	tests := []test{
		{
			input: fmt.Sprintf("%s/testdata/config_default.yaml", cwd),
			want: &Config{
				App: App{
					Timeout: consts.AppTimeout * time.Second,
				},
				Engine: Engine{
					Type: consts.EngineType,
				},
				Network: Network{
					Address:        consts.ServerAddress,
					MaxConnections: consts.MaxConnections,
				},
				Logger: Logger{
					Level:    consts.LogLevel,
					IsPretty: false,
				},
			},
		},
		{
			input: fmt.Sprintf("%s/testdata/config_custom.yaml", cwd),
			want: &Config{
				App: App{
					Timeout: 15 * time.Second,
				},
				Engine: Engine{
					Type: "custom",
				},
				Network: Network{
					Address:        "127.0.0.1:8080",
					MaxConnections: 10,
				},
				Logger: Logger{
					Level:    "error",
					IsPretty: true,
				},
			},
		},
	}

	for _, tc := range tests {
		got, err := NewConfig(tc.input)
		if err != nil {
			t.Fatalf("NewConfig(%s): %v", tc.input, err)
		}

		gotBytes, _ := json.Marshal(got)
		wantBytes, _ := json.Marshal(tc.want)
		if string(gotBytes) != string(wantBytes) {
			t.Errorf("NewConfig(%s): got %s, want %s", tc.input, gotBytes, wantBytes)
		}
	}
}

func TestNewConfig_InvalidPath(t *testing.T) {
	dir, err := ioutil.TempDir("", "config_test_invalid_path")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	tests := []string{
		"config_invalid_path.yaml",
		fmt.Sprintf("%s/invalid_file", dir),
	}

	for _, path := range tests {
		_, err := NewConfig(path)
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("NewConfig(%s): got unexpected error %v", path, err)
		}
	}
}

func TestNewConfig_InvalidFile(t *testing.T) {
	file, err := ioutil.TempFile("", "config_test_invalid_file")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString("invalid yaml")
	if err != nil {
		t.Fatalf("write to temp file: %v", err)
	}
	_ = file.Close()

	_, err = NewConfig(file.Name())
	if err == nil {
		t.Fatal("NewConfig(invalid_file): expected an error, got nil")
	}
}