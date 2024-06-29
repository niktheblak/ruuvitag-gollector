package scanner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type interval struct {
	scanner
}

func NewInterval(device string, peripherals map[string]string, exporters []exporter.Exporter, logger *slog.Logger) (Scanner, error) {
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

func NewIntervalWithOpts(cfg Config) (Scanner, error) {
	s := &interval{
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

// Scan scans and reports measurements at specified intervals
func (s *interval) Scan(ctx context.Context, scanInterval time.Duration) error {
	if scanInterval == 0 {
		return fmt.Errorf("scan interval must be greater than zero")
	}
	if len(s.exporters) == 0 {
		return fmt.Errorf("at least one exporter must be specified")
	}
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Scanning measurements", slog.Duration("interval", scanInterval))
	ticker := time.NewTicker(scanInterval)
	s.listen(ctx, ticker.C, scanInterval)
	ticker.Stop()
	return nil
}

func (s *interval) listen(ctx context.Context, ticks <-chan time.Time, scanTimeout time.Duration) {
	for {
		select {
		case <-ticks:
			scanCtx, cancel := context.WithTimeout(ctx, scanTimeout)
			s.doScan(scanCtx)
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func (s *interval) doScan(ctx context.Context) {
	meas := s.meas.Channel(ctx)
	s.doExport(ctx, meas)
}
