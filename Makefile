GOPATH := $(shell go env GOPATH)
SWAG := $(GOPATH)/bin/swag

.PHONY: build run test tidy fmt swagger swagger-install docs compose-up compose-down compose-logs

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./...

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

