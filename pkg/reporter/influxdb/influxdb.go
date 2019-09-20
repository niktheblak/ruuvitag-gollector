package influxdb

import (
	"context"
	"os"

	"github.com/influxdata/influxdb-client-go"
	"github.com/niktheblak/ruuvitag-gollector/pkg/reporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type influxdbReporter struct {
	client *influxdb.Client
}

func New() (reporter.Reporter, error) {
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
	return &influxdbReporter{
		client: client,
	}, nil
}

func (r *influxdbReporter) Name() string {
	return "InfluxDB"
}

func (r *influxdbReporter) Report(data ruuvitag.SensorData) error {
	ctx := context.Background()
	m := influxdb.NewRowMetric(map[string]interface{}{
		"temperature": data.Temperature,
		"humidity":    data.Humidity,
		"pressure":    data.Pressure,
	}, "ruuvitag", map[string]string{
		"mac":  data.DeviceID,
		"name": data.Name,
	}, data.Timestamp)
	_, err := r.client.Write(ctx, "", "", m)
	return err
}

func (r *influxdbReporter) Close() error {
	return r.client.Close()
}
