GOPATH := $(shell go env GOPATH)
SWAG := $(GOPATH)/bin/swag

.PHONY: build run test test-coverage tidy fmt swagger swagger-install docs compose-up compose-down compose-logs

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger: $(SWAG)
	$(SWAG) init -g cmd/server/main.go -o docs

docs: swagger

compose-up:
	podman compose up --build -d

compose-down:
	podman compose down

compose-logs:
	podman compose logs -f

$(SWAG):
	$(MAKE) swagger-install

