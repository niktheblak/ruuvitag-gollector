.PHONY: all build install

all:
	make build
	make install

build:
	go build -tags "influxdb postgresql gcp aws" -o ruuvitag-gollector main.go

install:
	cp ruuvitag-gollector ~/bin/

test:
	go test ./...

fmt:
	go fmt ./...