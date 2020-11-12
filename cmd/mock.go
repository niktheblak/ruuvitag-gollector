package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Send mock data to configured exporters",
	RunE: func(cmd *cobra.Command, args []string) error {
		sendMockMeasurement()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mockCmd)
}

func sendMockMeasurement() {
	ts := time.Now()
	var measurements []sensor.Data
	for addr, name := range peripherals {
		measurements = append(measurements, generateMockData(addr, name, ts))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, exporter := range exporters {
		logger.Info("Sending mock measurement to exporter", zap.String("exporter", exporter.Name()))
		for _, data := range measurements {
			if err := exporter.Export(ctx, data); err != nil {
				logger.Error("Failed to export measurement", zap.Error(err))
			}
		}
	}
}

func generateMockData(addr, name string, ts time.Time) sensor.Data {
	return sensor.Data{
		Addr:            addr,
		Name:            name,
		Temperature:     21.5,
		Humidity:        60,
		Pressure:        1002,
		BatteryVoltage:  2.755,
		AccelerationX:   0,
		AccelerationY:   0,
		AccelerationZ:   0,
		MovementCounter: 0,
		Timestamp:       ts,
	}
}
