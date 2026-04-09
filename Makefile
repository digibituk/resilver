BINARY := resilver
MODULE := github.com/digibituk/resilver
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test test-e2e test-all run lint clean build-pi build-pi64 build-linux

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/resilver

test:
	go test ./... -v -race

test-e2e:
	npx playwright test

test-all: test test-e2e

run: build
	./bin/$(BINARY)

lint:
	go vet ./...

build-pi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm ./cmd/resilver

build-pi64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 ./cmd/resilver

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 ./cmd/resilver

clean:
	rm -rf bin/
