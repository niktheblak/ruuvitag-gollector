#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  --network ruuvitag \
  influxdb:1.8 \
  influx \
  -host influxdb \
  -database ruuvitag \
  -username user \
  -password changeme
