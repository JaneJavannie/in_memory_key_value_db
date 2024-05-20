SERVER_FILE = cmd/main.go
CLIENT_FILE = client/main.go
OUTPUT_DIR = .

build_server:
	go build -o $(OUTPUT_DIR)/in_memory_db $(SERVER_FILE)

build_client:
	go build -o $(OUTPUT_DIR)/in_memory_db_client $(CLIENT_FILE)

test:
	go test ./... -v -cover