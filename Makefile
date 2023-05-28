# Name of the binary to build
BINARY_NAME=ditto

# Go source files
SRC=$(shell find . -name "*.go" -type f)

# Build the binary for the current platform
build:
	go build -o $(BINARY_NAME) ./cmd/ditto

build-debug:
	go build -tags=debug -o $(BINARY_NAME) ./cmd/ditto

# Build the binary for linux/amd64
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) ./cmd/ditto

# Build the Docker image
docker-build:
	docker build -t ditto:latest .

# Clean the project
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run the tests
test:
	go test -v ./...

# Format the source code
fmt:
	gofmt -w $(SRC)
