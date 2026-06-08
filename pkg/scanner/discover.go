package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"tinygo.org/x/bluetooth"
)

type Discover struct {
	ble    BLEScanner
	logger *slog.Logger
}

func NewDiscover(ble BLEScanner, logger *slog.Logger) *Discover {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Discover{
		ble:    ble,
		logger: logger,
	}
}

func (d *Discover) Discover(ctx context.Context) ([]string, error) {
	ch := make(chan string, 1024)
	err := d.ble.Scan(ctx, func(result bluetooth.ScanResult) {
		if !Filter(nil, result) {
			return
		}
		addr := result.Address.String()
		d.logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
		ch <- addr
	})
	close(ch)
	switch {
	case errors.Is(err, context.Canceled):
	case errors.Is(err, context.DeadlineExceeded):
	case err == nil:
	default:
		return nil, err
	}
	addrMap := make(map[string]bool)
	for addr := range ch {
		addrMap[strings.ToUpper(addr)] = true
	}
	return slices.Sorted(maps.Keys(addrMap)), nil
}
