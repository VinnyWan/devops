.PHONY: all build run test clean fmt vet swag openapi build-web help \
       lint test-race test-coverage docker-build pre-commit install-tools

PROJECT_NAME := devops-server
MAIN_FILE := cmd/server/main.go
WEB_DIR := web
COVERAGE_FILE := coverage.out
DOCKER_IMAGE := devops-server
DOCKER_TAG := latest

all: build

## Run: Run the server directly
run:
	go run $(MAIN_FILE)

## Build: Build the binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(PROJECT_NAME) $(MAIN_FILE)

## Build Web: Build the frontend assets
build-web:
	cd $(WEB_DIR) && npm install && npm run build

## Test: Run unit tests
test:
	go test ./... -v

## Test Race: Run tests with race detector
test-race:
	go test ./... -race -v

## Test Coverage: Run tests and generate coverage report
test-coverage:
	go test ./... -coverprofile=$(COVERAGE_FILE) -covermode=atomic
	go tool cover -func=$(COVERAGE_FILE)
	@echo "---"
	@echo "HTML 报告: go tool cover -html=$(COVERAGE_FILE) -o coverage.html"

## Lint: Run golangci-lint
lint:
	golangci-lint run ./...

## Fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## Vet: Analyze code for potential errors
vet:
	go vet ./...

## Swag: Generate Swagger documentation
swag:
	swag init -g $(MAIN_FILE) -o docs/swagger
	$(MAKE) openapi

## OpenAPI: Convert Swagger 2.0 docs to OpenAPI 3.0
openapi:
	go run ./scripts/openapi/convert.go

## Docker Build: Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Docker Run: Run Docker container
docker-run:
	docker run -p 8000:8000 --rm $(DOCKER_IMAGE):$(DOCKER_TAG)

## Pre Commit: Run all checks before committing
pre-commit: fmt vet lint test
	@echo "所有检查通过"

## Install Tools: Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

## Security Check: Run vulnerability scanner
security-check:
	govulncheck ./...

## Clean: Remove binaries and temp files
clean:
	rm -f $(PROJECT_NAME)
	rm -f $(COVERAGE_FILE)
	rm -f coverage.html

## Help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
