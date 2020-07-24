package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/evenminutes"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"go.uber.org/zap"
)

type Scanner struct {
	Exporters []exporter.Exporter
	Quit      chan int

	logger      *zap.Logger
	device      ble.Device
	peripherals map[string]string
	stopped     bool
	dev         DeviceCreator
	ble         BLEScanner
}

func New(logger *zap.Logger, peripherals map[string]string) *Scanner {
	peripheralMap := make(map[string]string)
	for addr, name := range peripherals {
		peripheralMap[ble.NewAddr(addr).String()] = name
	}
	return &Scanner{
		Quit:        make(chan int, 1),
		logger:      logger,
		peripherals: peripheralMap,
		dev:         defaultDeviceCreator{},
		ble:         defaultBLEScanner{},
	}
}

// ScanContinuously scans and reports measurements immediately as they are received
func (s *Scanner) ScanContinuously(ctx context.Context) {
	s.logger.Info("Listening for measurements")
	ctx, cancel := context.WithCancel(ctx)
	meas := s.Measurements(ctx)
	go s.doExportContinuously(ctx, meas)
	go func() {
		select {
		case <-s.Quit:
		case <-ctx.Done():
		}
		cancel()
		s.Stop()
	}()
}

// ScanWithInterval scans and reports measurements at specified intervals
func (s *Scanner) ScanWithInterval(ctx context.Context, scanInterval time.Duration) {
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

// ScanOnce scans all registered peripherals once and quits
func (s *Scanner) ScanOnce(ctx context.Context) error {
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	s.doScan(ctx)
	s.Stop()
	return nil
}

// Measurements creates a channel that will receive measurements read from all registered peripherals.
// The cancel function should be called after the client is done with receiving measurements or wishes
// to abort the scan.
func (s *Scanner) Measurements(ctx context.Context) chan sensor.Data {
	ch := make(chan sensor.Data, 128)
	go func() {
		err := s.ble.Scan(ctx, false, s.handler(ch), s.filter)
		switch err {
		case context.Canceled:
		case context.DeadlineExceeded:
		case nil:
		default:
			s.logger.Error("Scan failed", zap.Error(err))
		}
		close(ch)
	}()
	return ch
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
	meas := s.Measurements(ctx)
	done := make(chan int, 1)
	go s.doExport(ctx, meas, done)
	select {
	case <-done:
	case <-s.Quit:
		done <- 1
	}
	close(done)
}

func (s *Scanner) doExportContinuously(ctx context.Context, measurements chan sensor.Data) {
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

func (s *Scanner) handler(ch chan sensor.Data) func(ble.Advertisement) {
	return func(a ble.Advertisement) {
		addr := a.Addr().String()
		s.logger.Debug("Read sensor data from device", zap.String("addr", addr))
		data := a.ManufacturerData()
		sensorData, err := sensor.Parse(data)
		if err != nil {
			s.logInvalidData(data, err)
			return
		}
		sensorData.Addr = addr
		sensorData.Name = s.peripherals[addr]
		sensorData.Timestamp = time.Now()
		ch <- sensorData
	}
}

func (s *Scanner) filter(a ble.Advertisement) bool {
	if !sensor.IsRuuviTag(a.ManufacturerData()) {
		return false
	}
	if len(s.peripherals) == 0 {
		return true
	}
	_, ok := s.peripherals[a.Addr().String()]
	return ok
}

func (s *Scanner) export(ctx context.Context, m sensor.Data) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for _, e := range s.Exporters {
		s.logger.Info("Exporting measurement", zap.String("exporter", e.Name()))
		err := retry.Do(func() error {
			return e.Export(ctx, m)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scanner) logInvalidData(data []byte, err error) {
	var header []byte
	if len(data) >= 3 {
		header = data[:3]
	} else {
		header = data
	}
	s.logger.Error("Error while parsing RuuviTag data",
		zap.Int("len", len(data)),
		zap.Binary("header", header),
		zap.Error(err),
	)
}
