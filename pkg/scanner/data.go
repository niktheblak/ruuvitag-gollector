package scanner

import (
	"context"
	"log/slog"
	"time"

	commonsensor "github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/dewpoint"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
	"github.com/niktheblak/ruuvitag-gollector/pkg/wetbulb"
)

// Read reads sensor data from advertisement
func Read(a Advertisement) (sd commonsensor.Data, err error) {
	sd, _, err = read(a)
	return
}

func read(a Advertisement) (sd commonsensor.Data, dataFormat uint8, err error) {
	addr := NormalizeAddress(a.Address)
	data := a.RawManufacturerData(sensor.RuuviManufacturerID)
	sd, err = sensor.Parse(data)
	if err != nil {
		return
	}
	dataFormat = data[2]
	sd.Addr = addr
	sd.Timestamp = time.Now()
	sd.DewPoint, err = dewpoint.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity)
	if err != nil {
		// dew point calculation failed, dew point will not be available
		sd.DewPoint = 0
	}
	sd.WetBulb, err = wetbulb.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity, sd.Pressure)
	if err != nil {
		// web bulb temperature was out of range, wet bulb temperature will not be available
		sd.WetBulb = 0
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
