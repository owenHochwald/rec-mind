.PHONY: build run test clean docker-build docker-run dev

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=main

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/server/main.go

# Run the application locally
run:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/server/main.go
	./$(BINARY_NAME)

# Run without building binary
dev:
	$(GOCMD) run cmd/server/main.go

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Docker commands
docker-build:
	docker build -t my-go-api .

docker-run:
	docker run -p 8080:8080 my-go-api

docker-compose-up:
	docker-compose up --build

docker-compose-down:
	docker-compose down