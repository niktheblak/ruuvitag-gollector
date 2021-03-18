// +build mqtt

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/mqtt"
)

func init() {
	rootCmd.PersistentFlags().Bool("mqtt.enabled", false, "Publish measurements to a MQTT broker")
	rootCmd.PersistentFlags().String("mqtt.addr", "tcp://localhost:1883", "MQTT broker address with protocol (tcp or ssl), host and port")
	rootCmd.PersistentFlags().String("mqtt.client_id", "ruuvitag-gollector", "MQTT client id")
	rootCmd.PersistentFlags().String("mqtt.username", "", "MQTT username")
	rootCmd.PersistentFlags().String("mqtt.password", "", "MQTT password")
	rootCmd.PersistentFlags().String("mqtt.ca_file", "", "Path to a CA file, if TLS used")
	rootCmd.PersistentFlags().Bool("mqtt.auto_reconnect", false, "Enable auto reconnection if connection is lost")
	rootCmd.PersistentFlags().Int("mqtt.reconnect_interval", 60, "Sets the maximum time in seconds that will be waited between reconnection attempts")
}

func addMQTTExporter(exporters *[]exporter.Exporter) error {
	addr := viper.GetString("mqtt.addr")
	if addr == "" {
		return fmt.Errorf("MQTT broker address must be specified")
	}
	exporter, err := mqtt.New(mqtt.Config{
		Addr:              addr,
		ClientId:          viper.GetString("mqtt.client_id"),
		Username:          viper.GetString("mqtt.username"),
		Password:          viper.GetString("mqtt.password"),
		CaFile:            viper.GetString("mqtt.ca_file"),
		AutoReconnect:     viper.GetBool("mqtt.auto_reconnect"),
		ReconnectInterval: time.Duration(viper.GetInt("mqtt.reconnect_interval")) * time.Second,
	})
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exporter)
	return nil
}
