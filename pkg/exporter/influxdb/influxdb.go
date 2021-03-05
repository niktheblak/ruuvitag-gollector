// +build influxdb

package influxdb

import (
	"context"
	"fmt"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type influxdbExporter struct {
	client      influxdb2.Client
	writeAPI    api.WriteAPIBlocking
	measurement string
}

func New(cfg Config) exporter.Exporter {
	token := cfg.Token
	if token == "" {
		token = fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
	}
	client := influxdb2.NewClient(cfg.Addr, token)
	bucket := cfg.Bucket
	if bucket == "" {
		bucket = cfg.Database
	}
	writeAPI := client.WriteAPIBlocking(cfg.Org, bucket)
	return &influxdbExporter{
		client:      client,
		writeAPI:    writeAPI,
		measurement: cfg.Measurement,
	}
}

func (e *influxdbExporter) Name() string {
	return "InfluxDB"
}

func (e *influxdbExporter) Export(ctx context.Context, data sensor.Data) error {
	point := influxdb2.NewPoint(e.measurement, map[string]string{
		"mac":  strings.ToUpper(data.Addr),
		"name": data.Name,
	}, map[string]interface{}{
		"temperature":        data.Temperature,
		"humidity":           data.Humidity,
		"dew_point":          data.DewPoint,
		"pressure":           data.Pressure,
		"battery_voltage":    data.BatteryVoltage,
		"tx_power":           data.TxPower,
		"acceleration_x":     data.AccelerationX,
		"acceleration_y":     data.AccelerationY,
		"acceleration_z":     data.AccelerationZ,
		"movement_counter":   data.MovementCounter,
		"measurement_number": data.MeasurementNumber,
	}, data.Timestamp)
	return e.writeAPI.WritePoint(ctx, point)
}

func (e *influxdbExporter) Close() error {
	e.client.Close()
	return nil
}
