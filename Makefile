.PHONY: build clean test lint mock

all: templ build

templ:
	templ generate

build:
	go build -o lolmatchup.bin

test:
	go test ./... -cover

lint:
	gofmt -s -w .
	go vet ./...

mock:
	uv run cmd/mockserver/server.py

clean:
	rm -f lolmatchup.bin
