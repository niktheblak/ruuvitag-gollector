package influxdb

import (
	"context"
	"os"

	"github.com/influxdata/influxdb-client-go"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type influxdbExporter struct {
	client *influxdb.Client
}

func New() (exporter.Exporter, error) {
	url := os.Getenv("RUUVITAG_INFLUXDB_URL")
	username := os.Getenv("RUUVITAG_INFLUXDB_USERNAME")
	password := os.Getenv("RUUVITAG_INFLUXDB_PASSWORD")
	var opts []influxdb.Option
	if username != "" && password != "" {
		opts = append(opts, influxdb.WithUserAndPass(username, password))
	}
	client, err := influxdb.New(url, "", opts...)
	if err != nil {
		return nil, err
	}
	return &influxdbExporter{
		client: client,
	}, nil
}

func (e *influxdbExporter) Name() string {
	return "InfluxDB"
}

func (e *influxdbExporter) Export(ctx context.Context, data ruuvitag.SensorData) error {
	m := influxdb.NewRowMetric(map[string]interface{}{
		"temperature": data.Temperature,
		"humidity":    data.Humidity,
		"pressure":    data.Pressure,
	}, "ruuvitag", map[string]string{
		"mac":  data.DeviceID,
		"name": data.Name,
	}, data.Timestamp)
	_, err := e.client.Write(ctx, "", "", m)
	return err
}

func (e *influxdbExporter) Close() error {
	return e.client.Close()
}
