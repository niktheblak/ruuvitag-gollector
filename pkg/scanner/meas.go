package scanner

import (
	"context"
	"log/slog"

	"github.com/go-ble/ble"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

const BufferSize = 128

type Measurements struct {
	BLE         BLEScanner
	Peripherals map[string]string
	Logger      *slog.Logger
}

// Channel creates a channel that will receive measurements read from all registered peripherals.
// The cancel function should be called after the client is done with receiving measurements or wishes
// to abort the scan.
func (s *Measurements) Channel(ctx context.Context) chan sensor.Data {
	if s.Logger == nil {
		s.Logger = slog.Default()
	}
	ch := make(chan sensor.Data, BufferSize)
	go func() {
		err := s.BLE.Scan(ctx, true, func(a ble.Advertisement) {
			addr := a.Addr().String()
			s.Logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
			sensorData, err := Read(a)
			if err != nil {
				LogInvalidData(ctx, s.Logger, a.ManufacturerData(), err)
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
			s.Logger.LogAttrs(ctx, slog.LevelError, "Scan failed", slog.Any("error", err))
		}
		close(ch)
	}()
	return ch
}
