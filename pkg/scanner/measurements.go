package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"

	commonsensor "github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Measurements struct {
	BLE         BLEAdapter
	Peripherals map[string]string
	Logger      *slog.Logger
}

// Channel starts scanning and returns the measurements and the terminal scan
// result. Callers must cancel ctx when they no longer need measurements.
func (s *Measurements) Channel(ctx context.Context) (<-chan commonsensor.Data, <-chan error) {
	if s.Logger == nil {
		s.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	ch := make(chan commonsensor.Data)
	done := make(chan error, 1)
	go func() {
		defer close(ch)
		defer close(done)
		done <- s.scan(ctx, ch)
	}()
	return ch, done
}

func (s *Measurements) scan(ctx context.Context, ch chan<- commonsensor.Data) error {
	filter := Filter(s.Peripherals)
	err := s.BLE.Scan(ctx, func(a Advertisement) {
		if !filter(a) {
			return
		}
		addr := NormalizeAddress(a.Address)
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
		sensorData, err := Read(a)
		if err != nil {
			LogInvalidData(ctx, s.Logger, a.RawManufacturerData(sensor.RuuviManufacturerID), err)
			return
		}
		sensorData.Name = s.Peripherals[addr]
		select {
		case ch <- sensorData:
		case <-ctx.Done():
		}
	})
	switch {
	case errors.Is(err, context.Canceled):
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Context canceled", slog.Any("error", err))
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		s.Logger.LogAttrs(ctx, slog.LevelDebug, "Deadline exceeded", slog.Any("error", err))
		return nil
	case err == nil:
		return nil
	default:
		s.Logger.LogAttrs(ctx, slog.LevelError, "Scan failed", slog.Any("error", err))
		return err
	}
}
