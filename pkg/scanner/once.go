package scanner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type once struct {
	scanner
}

func NewOnce(device string, peripherals map[string]string, exporters []exporter.Exporter, logger *slog.Logger) (Scanner, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return NewOnceWithOpts(Config{
		Exporters:     exporters,
		DeviceName:    device,
		BLEScanner:    &defaultBLEScanner{},
		Peripherals:   peripherals,
		DeviceCreator: &defaultDeviceCreator{},
		Logger:        logger,
	})
}

func NewOnceWithOpts(cfg Config) (Scanner, error) {
	s := &once{
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

// Scan scans all registered peripherals once and quits
func (s *once) Scan(ctx context.Context, _ time.Duration) error {
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	if len(s.exporters) == 0 {
		return fmt.Errorf("at least one exporter must be specified")
	}
	meas := s.meas.Channel(ctx)
	s.doExport(ctx, meas)
	return nil
}
