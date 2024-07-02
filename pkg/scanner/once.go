package scanner

import (
	"context"
	"time"
)

type once struct {
	scanner
}

func NewOnce(cfg Config) (Scanner, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	s := &once{
		scanner: newScanner(cfg),
	}
	err := s.init(cfg.DeviceName)
	return s, err
}

// Scan scans all registered peripherals once and quits
func (s *once) Scan(ctx context.Context, _ time.Duration) error {
	meas := s.meas.Channel(ctx)
	s.doExport(ctx, meas)
	return nil
}
