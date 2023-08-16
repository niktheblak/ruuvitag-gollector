package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-ble/ble"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type OnceScanner struct {
	Exporters []exporter.Exporter

	logger      *slog.Logger
	device      ble.Device
	peripherals map[string]string
	dev         DeviceCreator
	meas        *Measurements
}

func NewOnce(logger *slog.Logger, peripherals map[string]string) *OnceScanner {
	bleScanner := defaultBLEScanner{}
	return &OnceScanner{
		logger:      logger,
		peripherals: peripherals,
		dev:         defaultDeviceCreator{},
		meas: &Measurements{
			BLE:         bleScanner,
			Peripherals: peripherals,
			Logger:      logger,
		},
	}
}

// Scan scans all registered peripherals once and quits
func (s *OnceScanner) Scan(ctx context.Context) error {
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	meas := s.meas.Channel(ctx)
	done := make(chan int, 1)
	go s.doExport(ctx, meas, done)
	select {
	case <-done:
	}
	return nil
}

func (s *OnceScanner) Close() {
	if s.device != nil {
		if err := s.device.Stop(); err != nil {
			s.logger.LogAttrs(nil, slog.LevelError, "Error while stopping device", slog.Any("error", err))
		}
	}
	for _, e := range s.Exporters {
		if err := e.Close(); err != nil {
			s.logger.LogAttrs(nil, slog.LevelError, "Failed to close exporter", slog.String("exporter", e.Name()), slog.Any("error", err))
		}
	}
}

func (s *OnceScanner) Init(device string) error {
	d, err := s.dev.NewDevice(device)
	if err != nil {
		return fmt.Errorf("failed to initialize device %s: %w", device, err)
	}
	s.device = d
	if len(s.peripherals) > 0 {
		s.logger.LogAttrs(nil, slog.LevelInfo, "Reading from peripherals", slog.Any("peripherals", s.peripherals))
	} else {
		s.logger.Info("Reading from all nearby BLE peripherals")
	}
	return nil
}

func (s *OnceScanner) doExport(ctx context.Context, measurements chan sensor.Data, done chan int) {
	seenPeripherals := make(map[string]bool)
	for {
		select {
		case m, ok := <-measurements:
			if !ok {
				done <- 1
				return
			}
			seenPeripherals[m.Addr] = true
			if err := s.export(ctx, m); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Failed to report measurement", slog.Any("error", err))
			}
			if len(s.peripherals) > 0 && ContainsKeys(s.peripherals, seenPeripherals) {
				done <- 1
				return
			}
		case <-ctx.Done():
			done <- 1
			return
		}
	}
}

func (s *OnceScanner) export(ctx context.Context, m sensor.Data) error {
	s.logger.LogAttrs(ctx, slog.LevelInfo, "Exporting measurement", slog.Any("data", m))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for _, e := range s.Exporters {
		if err := e.Export(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
