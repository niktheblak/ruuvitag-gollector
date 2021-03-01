#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  --name influxdb \
  --network ruuvitag \
  -p 8086:8086 \
  -e INFLUXDB_DB=ruuvitag \
  -e INFLUXDB_ADMIN_ENABLED=true \
  -e INFLUXDB_HTTP_AUTH_ENABLED=true \
  -e INFLUXDB_REPORTING_DISABLED=true \
  -e INFLUXDB_ADMIN_USER=admin \
  -e INFLUXDB_ADMIN_PASSWORD=Flux_Docker_Temp_Change_Me \
  -e INFLUXDB_USER=user \
  -e INFLUXDB_USER_PASSWORD=changeme \
  influxdb:1.8
