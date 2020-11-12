package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/go-ble/ble"
	"go.uber.org/zap"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type OnceScanner struct {
	Exporters []exporter.Exporter

	logger      *zap.Logger
	device      ble.Device
	peripherals map[string]string
	stopped     bool
	dev         DeviceCreator
	meas        *Measurements
}

func NewOnce(logger *zap.Logger, peripherals map[string]string) *OnceScanner {
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
	if s.stopped {
		return fmt.Errorf("scanner has already run")
	}
	meas := s.meas.Channel(ctx)
	done := make(chan int, 1)
	go s.doExport(ctx, meas, done)
	select {
	case <-done:
	}
	s.logger.Info("Stopping scanner")
	s.close()
	s.stopped = true
	return nil
}

func (s *OnceScanner) close() {
	if s.device != nil {
		if err := s.device.Stop(); err != nil {
			s.logger.Error("Error while stopping device", zap.Error(err))
		}
	}
	for _, e := range s.Exporters {
		if err := e.Close(); err != nil {
			s.logger.Error("Failed to close exporter", zap.String("exporter", e.Name()), zap.Error(err))
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
		s.logger.Info("Reading from peripherals", zap.Any("peripherals", s.peripherals))
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
				s.logger.Error("Failed to report measurement", zap.Error(err))
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
	s.logger.Info("Exporting measurement", zap.Any("data", m))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for _, e := range s.Exporters {
		if err := e.Export(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
