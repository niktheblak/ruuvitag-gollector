package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-ble/ble"
	"go.uber.org/zap"

	"github.com/niktheblak/ruuvitag-gollector/pkg/evenminutes"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Scanner struct {
	Exporters []exporter.Exporter
	Quit      chan int

	logger      *zap.Logger
	device      ble.Device
	peripherals map[string]string
	stopped     bool
	dev         DeviceCreator
	meas        *Measurements
}

func NewInterval(logger *zap.Logger, peripherals map[string]string) *Scanner {
	bleScanner := defaultBLEScanner{}
	return &Scanner{
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

// Scan scans and reports measurements at specified intervals
func (s *Scanner) Scan(ctx context.Context, scanInterval time.Duration) {
	if scanInterval == 0 {
		s.logger.Error("scan interval must be greater than zero")
		return
	}
	go func() {
		delay := evenminutes.Until(time.Now(), scanInterval)
		s.logger.Info("Sleeping", zap.String("until", time.Now().Add(delay).String()))
		firstRun := time.After(delay)
		select {
		case <-firstRun:
		case <-ctx.Done():
			return
		case <-s.Quit:
			return
		}
		s.logger.Info("Scanning measurements", zap.Duration("interval", scanInterval))
		ticker := time.NewTicker(scanInterval)
		firstCtx, cancel := context.WithTimeout(ctx, scanInterval)
		s.doScan(firstCtx)
		cancel()
		s.listen(ctx, ticker.C, scanInterval)
		ticker.Stop()
		s.Stop()
	}()
}

// Stop stops all running scans
func (s *Scanner) Stop() {
	if s.stopped {
		return
	}
	s.logger.Info("Stopping scanner")
	s.stopped = true
	s.Quit <- 1
}

// Close closes the scanner and frees allocated resources
func (s *Scanner) Close() {
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
func (s *Scanner) Init(device string) error {
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

func (s *Scanner) listen(ctx context.Context, ticks <-chan time.Time, scanTimeout time.Duration) {
	for {
		select {
		case <-ticks:
			ctx, cancel := context.WithTimeout(ctx, scanTimeout)
			s.doScan(ctx)
			cancel()
		case <-ctx.Done():
			return
		case <-s.Quit:
			return
		}
	}
}

func (s *Scanner) doScan(ctx context.Context) {
	meas := s.meas.Channel(ctx)
	done := make(chan int, 1)
	go s.doExport(ctx, meas, done)
	select {
	case <-done:
	case <-s.Quit:
		done <- 1
	}
	close(done)
}

func (s *Scanner) doExport(ctx context.Context, measurements chan sensor.Data, done chan int) {
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

func (s *Scanner) export(ctx context.Context, m sensor.Data) error {
	s.logger.Info("Exporting measurement", zap.Any("data", m))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for _, e := range s.Exporters {
		err := retry.Do(func() error {
			return e.Export(ctx, m)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
