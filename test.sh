#!/usr/bin/env bash

set -e

docker run -it -v "$(pwd):/go/src/app" ruuvitag-gollector:latest test ./...
