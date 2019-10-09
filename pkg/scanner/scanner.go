package scanner

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Scanner struct {
	Exporters    []exporter.Exporter
	logger       *log.Logger
	device       ble.Device
	quit         chan int
	measurements chan sensor.Data
	peripherals  map[string]string
	deviceImpl   string
	stopped      bool
	dev          DeviceCreator
	ble          BLEScanner
}

func New(logger *log.Logger, device string, peripherals map[string]string) (*Scanner, error) {
	return &Scanner{
		logger:       logger,
		quit:         make(chan int, 1),
		measurements: make(chan sensor.Data),
		peripherals:  peripherals,
		deviceImpl:   device,
		dev:          defaultDeviceCreator{},
		ble:          defaultBLEScanner{},
	}, nil
}

// ScanContinuously scans and reports measurements immediately as they are received
func (s *Scanner) ScanContinuously(ctx context.Context) error {
	if err := s.init(); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		s.logger.Println("Starting scan")
		err := s.ble.Scan(ctx, false, s.handle, s.filter)
		switch err {
		case context.Canceled:
			s.logger.Printf("Scan canceled")
		case nil:
			s.quit <- 1
		default:
			s.logger.Printf("Scan failed: %v", err)
			s.quit <- 1
		}
	}()
	go func() {
		s.logger.Println("Listening for measurements")
		for {
			select {
			case m := <-s.measurements:
				if err := s.export(ctx, m); err != nil {
					s.logger.Printf("Failed to report measurement: %v", err)
				}
			case <-s.quit:
				s.stopped = true
				cancel()
				return
			}
		}
	}()
	return nil
}

// ScanWithInterval scans and reports measurements at specified intervals
func (s *Scanner) ScanWithInterval(ctx context.Context, scanInterval time.Duration) error {
	if err := s.init(); err != nil {
		return err
	}
	ticker := time.NewTicker(scanInterval)
	go func() {
		s.logger.Printf("Scanning measurements every %v", scanInterval)
		for {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			select {
			case <-ticker.C:
				err := s.ble.Scan(ctx, false, s.handle, s.filter)
				switch err {
				case context.Canceled:
					s.logger.Printf("scan canceled")
				case context.DeadlineExceeded:
				case nil:
				default:
					s.logger.Printf("Scan failed: %v", err)
					s.quit <- 1
				}
			case <-s.quit:
				s.logger.Println("Scanner quitting")
				s.stopped = true
				cancel()
				ticker.Stop()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			case m := <-s.measurements:
				if err := s.export(ctx, m); err != nil {
					s.logger.Printf("Failed to report measurement: %v", err)
				}
			case <-s.quit:
				return
			}
		}
	}()
	return nil
}

// ScanOnce scans all registered peripherals once and quits
func (s *Scanner) ScanOnce(ctx context.Context) error {
	if len(s.peripherals) == 0 {
		return fmt.Errorf("at least one peripheral must be specified")
	}
	if err := s.init(); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		err := s.ble.Scan(ctx, false, s.handle, s.filter)
		switch err {
		case context.Canceled:
			s.logger.Println("scan canceled")
		case nil:
			s.quit <- 1
		default:
			s.logger.Printf("scan failed: %v", err)
			s.quit <- 1
		}
	}()
	seenPeripherals := make(map[string]bool)
	for {
		select {
		case m := <-s.measurements:
			seenPeripherals[m.Addr] = true
			if err := s.export(ctx, m); err != nil {
				s.logger.Printf("Failed to report measurement: %v", err)
			}
			if ContainsKeys(s.peripherals, seenPeripherals) {
				s.quit <- 1
			}
		case <-s.quit:
			s.stopped = true
			cancel()
			return nil
		}
	}
}

// Stop stops all running scans
func (s *Scanner) Stop() {
	s.logger.Println("Stopping")
	s.stopped = true
	s.quit <- 1
}

// Close closes the scanner and frees allocated resources
func (s *Scanner) Close() {
	if !s.stopped {
		s.Stop()
	}
	if s.device != nil {
		if err := s.device.Stop(); err != nil {
			s.logger.Println(err)
		}
	}
	for _, e := range s.Exporters {
		if err := e.Close(); err != nil {
			s.logger.Printf("Failed to close exporter %s: %v", e.Name(), err)
		}
	}
}

func (s *Scanner) init() error {
	d, err := s.dev.NewDevice(s.deviceImpl)
	if err != nil {
		return fmt.Errorf("failed to initialize device %s: %w", s.deviceImpl, err)
	}
	s.device = d
	if len(s.peripherals) > 0 {
		s.logger.Printf("Reading from peripherals %v", s.peripherals)
	} else {
		s.logger.Println("Reading from all nearby BLE peripherals")
	}
	return nil
}

func (s *Scanner) handle(a ble.Advertisement) {
	s.logger.Printf("Read sensor data from device %s:%v", a.LocalName(), a.Addr())
	data := a.ManufacturerData()
	sensorData, err := sensor.Parse(data)
	if err != nil {
		s.logInvalidData(data, err)
		return
	}
	addr := a.Addr().String()
	sensorData.Addr = strings.ToUpper(addr)
	sensorData.Name = s.peripherals[addr]
	sensorData.Timestamp = time.Now()
	s.measurements <- sensorData
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
	for _, e := range s.Exporters {
		s.logger.Printf("Exporting measurement to %v", e.Name())
		if err := e.Export(ctx, m); err != nil {
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
	s.logger.Printf("Error while parsing RuuviTag data (%d bytes) %v: %v", len(data), header, err)
}
