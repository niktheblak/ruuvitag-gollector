#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  --name postgres \
  --network ruuvitag \
  -p 5432:5432 \
  -e POSTGRES_USER=ruuvitag \
  -e POSTGRES_PASSWORD=changeme \
  postgres:latest
