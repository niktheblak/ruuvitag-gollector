package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	"tinygo.org/x/bluetooth"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

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
		s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	ch := make(chan sensor.Data)
	go s.scan(ctx, ch)
	return ch
}

func (s *Measurements) scan(ctx context.Context, ch chan sensor.Data) {
	err := s.BLE.Scan(ctx, func(result bluetooth.ScanResult) {
		if !Filter(s.Peripherals, result) {
			return
		}
		addr := strings.ToUpper(result.Address.String())
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
		for _, md := range result.ManufacturerData() {
			sensorData, err := Read(addr, md.Data)
			if err != nil {
				LogInvalidData(ctx, s.Logger, md.Data, err)
				return
			}
			sensorData.Name = s.Peripherals[addr]
			ch <- sensorData
		}
	})
	switch {
	case errors.Is(err, context.Canceled):
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Context canceled", slog.Any("error", err))
	case errors.Is(err, context.DeadlineExceeded):
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Deadline exceeded", slog.Any("error", err))
	case err == nil:
		// no error, ignore
	default:
		s.Logger.LogAttrs(ctx, slog.LevelError, "Scan failed", slog.Any("error", err))
	}
}
