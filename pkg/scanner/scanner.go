package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type Config struct {
	Exporters     []exporter.Exporter
	DeviceName    string
	BLEScanner    BLEScanner
	Peripherals   map[string]string
	DeviceCreator DeviceCreator
	Logger        *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		DeviceName:    "default",
		BLEScanner:    new(GoBLEScanner),
		DeviceCreator: new(GoBLEDeviceCreator),
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func Validate(cfg Config) error {
	if len(cfg.Exporters) == 0 {
		return fmt.Errorf("at least one exporter must be specified")
	}
	if len(cfg.Peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	return nil
}

type Scanner interface {
	io.Closer
	Scan(ctx context.Context, interval time.Duration) error
}

type scanner struct {
	exporters   []exporter.Exporter
	device      ble.Device
	peripherals map[string]string
	dev         DeviceCreator
	meas        *Measurements
	logger      *slog.Logger
}

func newScanner(cfg Config) scanner {
	return scanner{
		exporters:   cfg.Exporters,
		peripherals: cfg.Peripherals,
		dev:         cfg.DeviceCreator,
		logger:      cfg.Logger,
		meas: &Measurements{
			BLE:         cfg.BLEScanner,
			Peripherals: cfg.Peripherals,
			Logger:      cfg.Logger,
		},
	}
}

func (s *scanner) init(device string) error {
	d, err := s.dev.NewDevice(device)
	if err != nil {
		return fmt.Errorf("failed to initialize device %s: %w", device, err)
	}
	s.device = d
	if len(s.peripherals) > 0 {
		s.logger.LogAttrs(context.TODO(), slog.LevelInfo, "Reading from peripherals", slog.Any("peripherals", s.peripherals))
	} else {
		s.logger.Info("Reading from all nearby BLE peripherals")
	}
	return nil
}

func (s *scanner) Close() error {
	if s.device != nil {
		err := s.device.Stop()
		s.device = nil
		return err
	}
	return nil
}

func (s *scanner) doExport(ctx context.Context, measurements chan sensor.Data) {
	seenPeripherals := make(map[string]bool)
	for {
		select {
		case m, ok := <-measurements:
			if !ok {
				return
			}
			seenPeripherals[m.Addr] = true
			if err := s.export(ctx, m); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Failed to report measurement", slog.Any("error", err))
			}
			if len(s.peripherals) > 0 && containsKeys(s.peripherals, seenPeripherals) {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *scanner) export(ctx context.Context, m sensor.Data) error {
	if len(s.exporters) == 0 {
		return fmt.Errorf("no exporters available")
	}
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Exporting measurement", slog.Any("measurement", m))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	wg := new(sync.WaitGroup)
	wg.Add(len(s.exporters))
	errs := make(chan error, len(s.exporters))
	for _, e := range s.exporters {
		go func() {
			s.logger.LogAttrs(ctx, slog.LevelDebug, "Exporting measurement", slog.String("exporter", e.Name()))
			if err := e.Export(ctx, m); err != nil {
				errs <- err
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(errs)
	var allErrs []error
	for e := range errs {
		allErrs = append(allErrs, e)
	}
	return errors.Join(allErrs...)
}

func containsKeys[Map1 ~map[K]V1, K comparable, V1 any, Map2 ~map[K]V2, V2 any](m1 Map1, m2 Map2) bool {
	return maps.EqualFunc(m1, m2, func(_ V1, _ V2) bool {
		return true
	})
}
