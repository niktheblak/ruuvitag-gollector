#!/usr/bin/env bash

TAGS="influxdb postgresql gcp aws"

docker run \
  -it \
  --rm \
  -v "$(pwd):/go/src/app" \
  --network ruuvitag \
  ruuvitag-gollector:latest \
  run -tags "$TAGS" main.go mock --config ruuvitag-gollector.yaml
