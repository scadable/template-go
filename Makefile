# Makefile for template-go

.PHONY: setup runserver test lint

# Set up the local dev environment
setup:
	chmod +x ./bin/setup.sh
	./bin/setup.sh

# Run the Go application
runserver:
	go run ./cmd/template-go

# Run tests
test:
	go test ./...

# (Optional) Linting with golangci-lint
lint:
	golangci-lint run
