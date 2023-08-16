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

type ContinuousScanner struct {
	Exporters []exporter.Exporter
	Quit      chan int

	logger      *slog.Logger
	device      ble.Device
	peripherals map[string]string
	stopped     bool
	dev         DeviceCreator
	meas        *Measurements
}

func NewContinuous(logger *slog.Logger, peripherals map[string]string) *ContinuousScanner {
	bleScanner := defaultBLEScanner{}
	return &ContinuousScanner{
		Quit:        make(chan int, 1),
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

// Scan scans and reports measurements immediately as they are received
func (s *ContinuousScanner) Scan(ctx context.Context) {
	s.logger.Info("Listening for measurements")
	ctx, cancel := context.WithCancel(ctx)
	meas := s.meas.Channel(ctx)
	go s.exportContinuously(ctx, meas)
	go func() {
		select {
		case <-s.Quit:
		case <-ctx.Done():
		}
		cancel()
		s.Stop()
	}()
}

// Stop stops all running scans
func (s *ContinuousScanner) Stop() {
	if s.stopped {
		return
	}
	s.logger.Info("Stopping scanner")
	s.stopped = true
	s.Quit <- 1
}

// Close closes the scanner and frees allocated resources
func (s *ContinuousScanner) Close() {
	if !s.stopped {
		s.Stop()
	}
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

// Init initializes scanner using the given device
func (s *ContinuousScanner) Init(device string) error {
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

func (s *ContinuousScanner) exportContinuously(ctx context.Context, measurements chan sensor.Data) {
	for {
		select {
		case m, ok := <-measurements:
			if !ok {
				return
			}
			if err := s.export(ctx, m); err != nil {
				s.logger.LogAttrs(ctx, slog.LevelError, "Failed to report measurement", slog.Any("error", err))
			}
		case <-ctx.Done():
			return
		case <-s.Quit:
			return
		}
	}
}

func (s *ContinuousScanner) export(ctx context.Context, m sensor.Data) error {
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
