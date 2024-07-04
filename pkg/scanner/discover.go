package scanner

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"

	"github.com/go-ble/ble"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
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
	err := d.ble.Scan(ctx, true, func(a ble.Advertisement) {
		addr := a.Addr().String()
		d.logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
		ch <- addr
	}, func(a ble.Advertisement) bool {
		return sensor.IsRuuviTag(a.ManufacturerData())
	})
	close(ch)
	switch {
	case errors.Is(err, context.Canceled):
	case errors.Is(err, context.DeadlineExceeded):
	case err == nil:
	default:
		return nil, err
	}
	addrMap := make(map[string]any)
	for addr := range ch {
		addrMap[addr] = struct{}{}
	}
	var addrs []string
	for addr := range addrMap {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)
	return addrs, nil
}
