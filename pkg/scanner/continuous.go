package scanner

import (
	"context"
	"log/slog"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

type continuous struct {
	scanner
}

func NewContinuous(cfg Config) (Scanner, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	s := &continuous{
		scanner: newScanner(cfg),
	}
	err := s.init(cfg.DeviceName)
	return s, err
}

// Scan scans and reports measurements immediately as they are received
func (s *continuous) Scan(ctx context.Context, _ time.Duration) error {
	s.logger.Info("Listening for measurements")
	meas := s.meas.Channel(ctx)
	s.exportContinuously(ctx, meas)
	return nil
}

func (s *continuous) exportContinuously(ctx context.Context, measurements chan sensor.Data) {
	for {
		select {
		case m, ok := <-measurements:
			if !ok {
				return
			}
			if err := s.export(ctx, m); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Failed to report measurement", slog.Any("error", err))
			}
		case <-ctx.Done():
			return
		}
	}
}
