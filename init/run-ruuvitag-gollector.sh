#!/bin/bash

sudo /home/pi/bin/ruuvitag-gollector \
  --config /home/pi/ruuvitag-gollector/configs/pi.toml \
  -d \
  --scan_interval 5m \
  --influxdb \
  --pubsub \
  --stackdriver