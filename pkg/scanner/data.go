package scanner

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-ble/ble"
	commonsensor "github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/dewpoint"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
	"github.com/niktheblak/ruuvitag-gollector/pkg/wetbulb"
)

// Read reads sensor data from advertisement
func Read(a ble.Advertisement) (sd commonsensor.Data, err error) {
	addr := a.Addr().String()
	data := a.ManufacturerData()
	sd, err = sensor.Parse(data)
	if err != nil {
		return
	}
	sd.Addr = addr
	sd.Timestamp = time.Now()
	sd.DewPoint, err = dewpoint.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity)
	if err != nil {
		return
	}
	sd.WetBulb, err = wetbulb.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity)
	if err != nil {
		return
	}
	return
}

// LogInvalidData logs invalid BLE advertisement data
func LogInvalidData(ctx context.Context, logger *slog.Logger, data []byte, err error) {
	var header []byte
	if len(data) >= 3 {
		header = data[:3]
	} else {
		header = data
	}
	logger.LogAttrs(ctx, slog.LevelError, "Error while parsing RuuviTag data",
		slog.Int("len", len(data)),
		slog.Any("header", header),
		slog.Any("error", err),
	)
}
