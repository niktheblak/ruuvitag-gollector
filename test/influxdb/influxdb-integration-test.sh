#!/usr/bin/env bash

go test -tags influxdb,integration_test github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb
