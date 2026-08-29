package scanner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
)

type Discover struct {
	factory AdapterFactory
	adapter BLEAdapter
	logger  *slog.Logger
}

func NewDiscover(device string, factory AdapterFactory, logger *slog.Logger) (*Discover, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	d := &Discover{
		factory: factory,
		logger:  logger,
	}
	if err := d.init(device); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Discover) init(device string) error {
	adapter, err := d.factory.NewAdapter(device)
	if err != nil {
		return fmt.Errorf("failed to initialize Bluetooth adapter %s: %w", device, err)
	}
	d.adapter = adapter
	return nil
}

func (d *Discover) Discover(ctx context.Context) ([]string, error) {
	addrMap := make(map[string]bool)
	filter := Filter(nil)
	err := d.adapter.Scan(ctx, func(a Advertisement) {
		if !filter(a) {
			return
		}
		addr := NormalizeAddress(a.Address)
		d.logger.LogAttrs(ctx, slog.LevelDebug, "Read sensor data from device", slog.String("addr", addr))
		addrMap[strings.ToUpper(addr)] = true
	})
	switch {
	case errors.Is(err, context.Canceled):
	case errors.Is(err, context.DeadlineExceeded):
	case err == nil:
	default:
		return nil, err
	}
	return slices.Sorted(maps.Keys(addrMap)), nil
}

func (d *Discover) Close() error {
	if d.adapter != nil {
		err := d.adapter.Close()
		d.adapter = nil
		return err
	}
	return nil
}
