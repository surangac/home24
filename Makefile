.PHONY: build run test docker-build docker-run clean

# Build variables
BINARY_NAME=analyzer
MAIN_PATH=./cmd/analyzer

# Build the application
build:
	go build -o ${BINARY_NAME} ${MAIN_PATH}

# Run the application
run:
	go run ${MAIN_PATH}

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	go clean
	rm -f ${BINARY_NAME}

# Build docker image
docker-build:
	docker build -t ${BINARY_NAME} .

# Run docker container
docker-run:
	docker run -p 8080:8080 ${BINARY_NAME}

# Default target
all: build 