package influxdb

import (
	"context"

	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type influxdbExporter struct {
	client      influx.Client
	database    string
	measurement string
}

type Config struct {
	URL         string
	Token       string
	Database    string
	Measurement string
	Username    string
	Password    string
}

func New(cfg Config) (exporter.Exporter, error) {
	client, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     cfg.URL,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return &influxdbExporter{
		client:      client,
		database:    cfg.Database,
		measurement: cfg.Measurement,
	}, nil
}

func (e *influxdbExporter) Name() string {
	return "InfluxDB"
}

func (e *influxdbExporter) Export(ctx context.Context, data sensor.Data) error {
	conf := influx.BatchPointsConfig{
		Database: e.database,
	}
	bp, err := influx.NewBatchPoints(conf)
	if err != nil {
		return err
	}
	point, err := influx.NewPoint(e.measurement, map[string]string{
		"mac":  data.Addr,
		"name": data.Name,
	}, map[string]interface{}{
		"temperature": data.Temperature,
		"humidity":    data.Humidity,
		"pressure":    float64(data.Pressure),
	}, data.Timestamp)
	bp.AddPoint(point)
	return e.client.Write(bp)
}

func (e *influxdbExporter) Close() error {
	return e.client.Close()
}
