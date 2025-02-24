.PHONY: build clean

all: templ build

templ: 
	templ generate

build:
	go build -o lolmatchup.bin

clean: 
	rm lolmatchup
