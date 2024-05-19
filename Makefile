MAIN_FILE = cmd/main.go
OUTPUT_DIR = .

build:
	go build -o $(OUTPUT_DIR)/in_memory_db $(MAIN_FILE)

test:
	go test ./... -v -cover