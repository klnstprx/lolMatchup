.PHONY: build clean

templ: 
	templ generate

build: templ
	go build -o lolmatchup

clean: 
	rm lolmatchup
