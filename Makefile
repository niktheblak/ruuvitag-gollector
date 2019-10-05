.PHONY: all build install

all:
	make build
	make install

build:
	go build -o ruuvitag-gollector cmd/collector/main.go

install:
	cp ruuvitag-gollector ~/bin/

test:
	go test ./...

fmt:
	go fmt ./...