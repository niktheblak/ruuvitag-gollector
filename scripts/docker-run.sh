#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  -e CONSOLE=true \
  -v "$(pwd):/go/src/app" \
  ruuvitag-gollector:latest \
  run main.go mock --config ruuvitag-gollector.yaml