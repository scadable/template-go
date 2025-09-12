# Makefile for template-go

.PHONY: setup runserver test lint swagger

EXCLUDE_DIRS := $(shell yq '.exclude[]' tests/config.yaml | paste -sd '|' -)

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


coverage:
	go test -coverprofile=coverage.out $(shell go list ./... | grep -vE "($(EXCLUDE_DIRS))")
	go tool cover -func=coverage.out
	rm coverage.out
	#go tool cover -html=coverage.out


# (Optional) Linting with golangci-lint
lint:
	golangci-lint run

# Generate Swagger docs
swagger:
	swag init -g cmd/template-go/main.go -o ./docs
