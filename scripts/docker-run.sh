#!/usr/bin/env bash

TAGS="influxdb postgres gcp aws mqtt"

docker run \
  -it \
  --rm \
  -v "$(pwd):/go/src/app" \
  --network ruuvitag \
  ruuvitag-gollector:latest \
  run -tags "$TAGS" main.go "$@"
