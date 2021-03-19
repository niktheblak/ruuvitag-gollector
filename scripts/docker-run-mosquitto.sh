#!/usr/bin/env bash

mosquitto_config=$(mktemp)
echo "listener 1883
allow_anonymous true" > "$mosquitto_config"

docker run \
  -it \
  --rm \
  --name mosquitto \
  --network ruuvitag \
  -p 1883:1883 \
  -v "$mosquitto_config":/mosquitto/config/mosquitto.conf\
  eclipse-mosquitto:2.0.8

rm "$mosquitto_config"
