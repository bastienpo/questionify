.PHONY: build run test tidy audit clean help

build:
	@echo "Building api..."
	go build -o api cmd/api/main.go

run:
	go run cmd/api/main.go

watch:
	@if command -v air > /dev/null; then \
		air; \
		echo "Watching...";\
	else \
		echo "air is not installed. Exiting..."; \
		exit 1; \
	fi

test:
	@echo "Running tests..."
	go test -v -race ./...

tidy:
	@echo "Formatting the project..."
	go fmt ./...
	@echo "Tidying up the go.mod file..."
	go mod tidy
	@echo "Verifying and vendoring dependencies..."
	go mod verify
	go mod vendor

audit:
	@echo "Checking module dependencies"
	go mod tidy -diff
	@echo "Verifying module dependencies"
	go mod verify
	@echo "Vetting code..."
	go vet ./...
	@if command -v staticcheck > /dev/null; then \
		staticcheck ./...; \
	else \
		echo "staticcheck is not installed. Skipping..."; \
	fi

clean:
	@echo "Cleaning up..."
	rm -rf api tmp

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build     Build the api"
	@echo "  run       Run the api"
	@echo "  test      Run the tests"
	@echo "  clean     Clean the build"
