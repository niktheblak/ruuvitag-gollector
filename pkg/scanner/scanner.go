package scanner

import (
	"strings"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/pkg/errors"
)

type Scanner struct {
	Exporters     []exporter.Exporter
	logger        *log.Logger
	device        ble.Device
	quit          chan int
	measurements  chan sensor.Data
	peripherals   map[string]string
	deviceImpl    string
	dev           DeviceCreator
	ble           BLEScanner
}

type DeviceCreator interface {
	NewDevice(impl string) (ble.Device, error)
}

type defaultDeviceCreator struct {
}

func (c defaultDeviceCreator) NewDevice(impl string) (ble.Device, error) {
	d, err := dev.NewDevice(impl)
	if err != nil {
		return nil, err
	}
	ble.SetDefaultDevice(d)
	return d, nil
}

type BLEScanner interface {
	Scan(ctx context.Context, allowDup bool, h ble.AdvHandler, f ble.AdvFilter) error
}

type defaultBLEScanner struct {
}

func (s defaultBLEScanner) Scan(ctx context.Context, allowDup bool, h ble.AdvHandler, f ble.AdvFilter) error {
	return ble.Scan(ctx, allowDup, h, f)
}

func New(logger *log.Logger, device string, peripherals map[string]string) (*Scanner, error) {
	return &Scanner{
		logger:        logger,
		quit:          make(chan int),
		measurements:  make(chan sensor.Data),
		peripherals:   peripherals,
		deviceImpl:    device,
		dev:           defaultDeviceCreator{},
		ble:           defaultBLEScanner{},
	}, nil
}

func (s *Scanner) Start(ctx context.Context) error {
	if err := s.init(); err != nil {
		return err
	}
	go func() {
		err := s.scan(ctx, func(m sensor.Data) {
			if err := s.doExport(ctx, m); err != nil {
				s.logger.Printf("Failed to report measurement: %v", err)
			}
		})
		if err != nil {
			s.logger.Println(err)
		}
		s.logger.Println("Scanner quitting")
	}()
	return nil
}

func (s *Scanner) ScanOnce(ctx context.Context) (err error) {
	if err = s.init(); err != nil {
		return err
	}
	seenPeripherals := make(map[string]bool)
	go func() {
		err = s.scan(ctx, func(m sensor.Data) {
			seenPeripherals[m.Addr] = true
			if err := s.doExport(ctx, m); err != nil {
				s.logger.Printf("Failed to report measurement: %v", err)
			}
			if ContainsKeys(s.peripherals, seenPeripherals) {
				s.quit <- 1
			}
		})
	}()
	<-s.quit
	return
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

func (s *Scanner) Stop() {
	s.quit <- 1
}

func (s *Scanner) Close() {
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

func (s *Scanner) scan(ctx context.Context, f func(sensor.Data)) (err error) {
	go func() {
		if err = s.ble.Scan(ctx, false, s.handle, s.filter); err != nil {
			switch errors.Cause(err) {
			case context.Canceled:
				err = fmt.Errorf("scan canceled")
			default:
				err = fmt.Errorf("scan failed: %w", err)
			}
			s.quit <- 1
		}
	}()
	for {
		select {
		case <-s.quit:
			return
		case <-ctx.Done():
			s.quit <- 1
			return
		case m := <-s.measurements:
			f(m)
		}
	}
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

func (s *Scanner) doExport(ctx context.Context, m sensor.Data) error {
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

func ContainsKeys(a map[string]string, b map[string]bool) bool {
	for k := range a {
		_, ok := b[k]
		if !ok {
			return false
		}
	}
	return true
}
