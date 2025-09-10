# Makefile for template-go

.PHONY: runserver test lint

# Run the Go application
runserver:
	go run ./cmd/template-go

# Run tests
test:
	go test ./...

# (Optional) Linting with golangci-lint
lint:
	golangci-lint run
