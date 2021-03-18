TAGS = influxdb postgres gcp aws mqtt

.PHONY: all build install

all:
	make build
	make install

build:
	go build -tags "$(TAGS)" -o ruuvitag-gollector main.go

install:
	cp ruuvitag-gollector ~/bin/

test:
	go test -tags "$(TAGS)" ./...

fmt:
	go fmt ./...