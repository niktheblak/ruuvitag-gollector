package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type interval struct {
	scanner
}

func NewInterval(cfg Config) (Scanner, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	s := &interval{
		scanner: newScanner(cfg),
	}
	err := s.init(cfg.DeviceName)
	return s, err
}

// Scan scans and reports measurements at specified intervals
func (s *interval) Scan(ctx context.Context, scanInterval time.Duration) error {
	if scanInterval == 0 {
		return fmt.Errorf("scan interval must be greater than zero")
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
