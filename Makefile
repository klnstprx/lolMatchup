.PHONY: build clean test lint

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

clean:
	rm -f lolmatchup.bin
