#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  -v "$(pwd):/go/src/app" \
  ruuvitag-gollector:latest \
  test \
  -tags "influxdb postgresql gcp aws" \
  ./...