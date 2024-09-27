//go:build mqtt

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/mqtt"
)

func createMQTTExporter(cfg map[string]any) (exporter.Exporter, error) {
	addr := cast.ToString(cfg["addr"])
	if addr == "" {
		return nil, fmt.Errorf("MQTT broker address must be specified")
	}
	return mqtt.New(mqtt.Config{
		Addr:              addr,
		ClientId:          cast.ToString(cfg["client_id"]),
		Username:          cast.ToString(cfg["username"]),
		Password:          cast.ToString(cfg["password"]),
		CaFile:            cast.ToString(cfg["ca_file"]),
		AutoReconnect:     cast.ToBool(cfg["auto_reconnect"]),
		ReconnectInterval: time.Duration(cast.ToInt(cfg["reconnect_interval"])) * time.Second,
	})
}
