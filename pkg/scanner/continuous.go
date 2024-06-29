package scanner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type continuous struct {
	scanner
}

func NewContinuous(device string, peripherals map[string]string, exporters []exporter.Exporter, logger *slog.Logger) (Scanner, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return NewContinuousWithOpts(Config{
		Exporters:     exporters,
		DeviceName:    device,
		BLEScanner:    &defaultBLEScanner{},
		Peripherals:   peripherals,
		DeviceCreator: &defaultDeviceCreator{},
		Logger:        logger,
	})
}

func NewContinuousWithOpts(cfg Config) (Scanner, error) {
	s := &continuous{
		scanner: scanner{
			exporters:   cfg.Exporters,
			peripherals: cfg.Peripherals,
			dev:         cfg.DeviceCreator,
			logger:      cfg.Logger,
			meas: &Measurements{
				BLE:         cfg.BLEScanner,
				Peripherals: cfg.Peripherals,
				Logger:      cfg.Logger,
			},
		},
	}
	err := s.init(cfg.DeviceName)
	return s, err
}

// Scan scans and reports measurements immediately as they are received
func (s *continuous) Scan(ctx context.Context, _ time.Duration) error {
	if len(s.exporters) == 0 {
		return fmt.Errorf("at least one exporter must be specified")
	}
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
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
