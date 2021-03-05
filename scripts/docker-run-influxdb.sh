#!/usr/bin/env bash

docker run \
  -it \
  --rm \
  --name influxdb \
  --network ruuvitag \
  -p 8086:8086 \
  -e DOCKER_INFLUXDB_INIT_MODE=setup \
  -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
  -e DOCKER_INFLUXDB_INIT_PASSWORD=ChangeMeAdminPassword \
  -e DOCKER_INFLUXDB_INIT_ORG=bitnik \
  -e DOCKER_INFLUXDB_INIT_BUCKET=ruuvitag \
  -e DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=ChangeMeAdminToken \
  -e INFLUXDB_REPORTING_DISABLED=true \
  influxdb:2.0
