package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Discover struct {
	ble    BLEScanner
	dev    DeviceCreator
	device ble.Device
	logger *slog.Logger
}

func NewDiscover(device string, ble BLEScanner, dev DeviceCreator, logger *slog.Logger) (*Discover, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	d := &Discover{
		ble:    ble,
		dev:    dev,
		logger: logger,
	}
	if err := d.init(device); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Discover) init(device string) error {
	dev, err := d.dev.NewDevice(device)
	if err != nil {
		return fmt.Errorf("failed to initialize device %s: %w", device, err)
	}
	d.device = dev
	return nil
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
	addrMap := make(map[string]bool)
	for addr := range ch {
		addrMap[strings.ToUpper(addr)] = true
	}
	return slices.Sorted(maps.Keys(addrMap)), nil
}

func (d *Discover) Close() error {
	if d.device != nil {
		err := d.device.Stop()
		d.device = nil
		return err
	}
	return nil
}
