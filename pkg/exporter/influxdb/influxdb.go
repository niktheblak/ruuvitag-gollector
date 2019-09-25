package influxdb

import (
	"context"

	"github.com/influxdata/influxdb-client-go"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type influxdbExporter struct {
	client *influxdb.Client
}

type Config struct {
	URL      string
	Token    string
	Username string
	Password string
}

func New(cfg Config) (exporter.Exporter, error) {
	var opts []influxdb.Option
	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, influxdb.WithUserAndPass(cfg.Username, cfg.Password))
	}
	client, err := influxdb.New(cfg.URL, cfg.Token, opts...)
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

func (e *influxdbExporter) Export(ctx context.Context, data sensor.Data) error {
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
