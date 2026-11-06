//go:build mqtt

package cmd

import (
	"context"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/mqtt"
)

func createMQTTExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exporterInitTimeout)
	defer cancel()
	// plain numbers are seconds, strings such as "30s" are parsed as durations
	reconnectInterval, err := configDuration(cfg["reconnect_interval"])
	if err != nil {
		return nil, err
	}
	// reconnect automatically unless explicitly disabled
	autoReconnect := true
	if v, ok := cfg["auto_reconnect"]; ok && v != nil {
		autoReconnect = cast.ToBool(v)
	}
	return mqtt.New(ctx, name, mqtt.Config{
		Addr:              cast.ToString(cfg["addr"]),
		ClientId:          cast.ToString(cfg["client_id"]),
		Username:          cast.ToString(cfg["username"]),
		Password:          cast.ToString(cfg["password"]),
		CaFile:            cast.ToString(cfg["ca_file"]),
		AutoReconnect:     autoReconnect,
		ReconnectInterval: reconnectInterval,
		Columns:           columns,
		Logger:            logger,
	})
}
