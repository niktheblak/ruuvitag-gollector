package scanner

import (
	"context"

	"github.com/go-ble/ble"
)

type BLEScanner interface {
	Scan(ctx context.Context, allowDup bool, h ble.AdvHandler, f ble.AdvFilter) error
}

type GoBLEScanner struct {
}

func (s *GoBLEScanner) Scan(ctx context.Context, allowDup bool, h ble.AdvHandler, f ble.AdvFilter) error {
	return ble.Scan(ctx, allowDup, h, f)
}
