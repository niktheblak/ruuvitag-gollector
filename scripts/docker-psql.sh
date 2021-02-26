#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  --network ruuvitag \
  -e PGPASSWORD=changeme \
  postgres:latest \
  psql \
  -h postgres \
  -U ruuvitag
