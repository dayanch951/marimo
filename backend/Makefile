.PHONY: proto run build clean

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/proto/auth.proto

# Run the server
run:
	go run cmd/server/main.go

# Build the server
build:
	go build -o bin/server cmd/server/main.go

# Download dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/
