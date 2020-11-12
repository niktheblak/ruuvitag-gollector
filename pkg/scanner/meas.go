package scanner

import (
	"context"

	"github.com/go-ble/ble"
	"go.uber.org/zap"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

const BufferSize = 128

type Measurements struct {
	BLE         BLEScanner
	Peripherals map[string]string
	Logger      *zap.Logger
}

// Channel creates a channel that will receive measurements read from all registered peripherals.
// The cancel function should be called after the client is done with receiving measurements or wishes
// to abort the scan.
func (s *Measurements) Channel(ctx context.Context) chan sensor.Data {
	if s.Logger == nil {
		s.Logger = zap.NewNop()
	}
	ch := make(chan sensor.Data, BufferSize)
	go func() {
		err := s.BLE.Scan(ctx, true, func(a ble.Advertisement) {
			addr := a.Addr().String()
			s.Logger.Debug("Read sensor data from device", zap.String("addr", addr))
			sensorData, err := Read(a)
			if err != nil {
				LogInvalidData(s.Logger, a.ManufacturerData(), err)
				return
			}
			sensorData.Name = s.Peripherals[addr]
			ch <- sensorData
		}, Filter(s.Peripherals))
		switch err {
		case context.Canceled:
		case context.DeadlineExceeded:
		case nil:
		default:
			s.Logger.Error("Scan failed", zap.Error(err))
		}
		close(ch)
	}()
	return ch
}
