package scanner

import (
	"time"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/dewpoint"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
	"go.uber.org/zap"
)

// Read reads sensor data from advertisement
func Read(a ble.Advertisement) (sd sensor.Data, err error) {
	addr := a.Addr().String()
	data := a.ManufacturerData()
	sd, err = sensor.Parse(data)
	sd.Addr = addr
	sd.Timestamp = time.Now()
	sd.DewPoint, _ = dewpoint.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity)
	return
}

// LogInvalidData logs invalid BLE advertisement data
func LogInvalidData(logger *zap.Logger, data []byte, err error) {
	var header []byte
	if len(data) >= 3 {
		header = data[:3]
	} else {
		header = data
	}
	logger.Error("Error while parsing RuuviTag data",
		zap.Int("len", len(data)),
		zap.Binary("header", header),
		zap.Error(err),
	)
}
