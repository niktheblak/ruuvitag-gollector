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

type ContinuousScanner struct {
	Exporters []exporter.Exporter
	Quit      chan int

	logger      *zap.Logger
	device      ble.Device
	peripherals map[string]string
	stopped     bool
	dev         DeviceCreator
	meas        *Measurements
}

func NewContinuous(logger *zap.Logger, peripherals map[string]string) *ContinuousScanner {
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
			s.logger.Error("Error while stopping device", zap.Error(err))
		}
	}
	for _, e := range s.Exporters {
		if err := e.Close(); err != nil {
			s.logger.Error("Failed to close exporter", zap.String("exporter", e.Name()), zap.Error(err))
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
		s.logger.Info("Reading from peripherals", zap.Any("peripherals", s.peripherals))
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
				s.logger.Error("Failed to report measurement", zap.Error(err))
			}
		case <-ctx.Done():
			return
		case <-s.Quit:
			return
		}
	}
}

func (s *ContinuousScanner) export(ctx context.Context, m sensor.Data) error {
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
