SERVER_FILE = cmd/main.go
CLIENT_FILE = client/main.go client/handler.go
OUTPUT_DIR = bin/

build_server:
	go build -o $(OUTPUT_DIR)/server $(SERVER_FILE)

build_client:
	go build -o $(OUTPUT_DIR)/client $(CLIENT_FILE)

build: build_server build_client


test:
	go test ./... -coverprofile cover.out && go tool cover -func cover.out

