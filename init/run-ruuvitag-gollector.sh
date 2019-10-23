#!/bin/bash

sudo GOOGLE_APPLICATION_CREDENTIALS=/home/pi/ruuvitag-firestore-bf4b3971ddbc.json /home/pi/bin/ruuvitag-gollector \
  --config /home/pi/ruuvitag-gollector/configs/pi.toml \
  -d \
  --scan_interval 5m \
  --influxdb \
  --pubsub \
  --stackdriver