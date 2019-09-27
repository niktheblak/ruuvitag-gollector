package scanner

import (
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
	SleepInterval time.Duration
	Exporters     []exporter.Exporter
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

func New(reportingInterval time.Duration, device string, peripherals map[string]string) (*Scanner, error) {
	return &Scanner{
		SleepInterval: reportingInterval,
		quit:          make(chan int),
		measurements:  make(chan sensor.Data),
		peripherals:   peripherals,
		deviceImpl:    device,
		dev:           defaultDeviceCreator{},
		ble:           defaultBLEScanner{},
	}, nil
}

func (s *Scanner) Start(ctx context.Context) error {
	d, err := s.dev.NewDevice(s.deviceImpl)
	if err != nil {
		return fmt.Errorf("failed to initialize device %s: %w", s.deviceImpl, err)
	}
	s.device = d
	if len(s.peripherals) > 0 {
		log.Printf("Reading from peripherals %v", s.peripherals)
	} else {
		log.Println("Reading from all nearby BLE peripherals")
	}
	go s.scan()
	go s.exportMeasurements(ctx)
	return nil
}

func (s *Scanner) Stop() {
	s.quit <- 1
	if err := s.device.Stop(); err != nil {
		log.Println(err)
	}
}

func (s *Scanner) scan() {
	log.Println("Scanner starting")
	timer := time.NewTimer(s.SleepInterval)
	defer timer.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	s.doScan(ctx)
	cancel()
	ctx, cancel = context.WithCancel(context.Background())
	for {
		select {
		case <-timer.C:
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			go s.doScan(ctx)
		case <-s.quit:
			log.Println("Scanner quitting")
			cancel()
			return
		}
	}
}

func (s *Scanner) doScan(ctx context.Context) {
	if err := s.ble.Scan(ctx, false, s.handle, s.filter); err != nil {
		switch errors.Cause(err) {
		case nil:
		case context.DeadlineExceeded:
			// Nothing found during scan window
		case context.Canceled:
			log.Println("Scan canceled")
		default:
			log.Printf("Scan failed: %v", err)
		}
	}
}

func (s *Scanner) handle(a ble.Advertisement) {
	log.Printf("Read sensor data from device %s:%v", a.LocalName(), a.Addr())
	data := a.ManufacturerData()
	sensorData, err := sensor.Parse(data)
	if err != nil {
		logInvalidData(data, err)
		return
	}
	sensorData.Addr = a.Addr().String()
	sensorData.Name = s.peripherals[sensorData.Addr]
	sensorData.Timestamp = time.Now()
	s.measurements <- sensorData
}

func logInvalidData(data []byte, err error) {
	var header []byte
	if len(data) >= 3 {
		header = data[:3]
	} else {
		header = data
	}
	log.Printf("Error while parsing RuuviTag data (%d bytes) %v: %v", len(data), header, err)
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

func (s *Scanner) exportMeasurements(ctx context.Context) {
	for {
		select {
		case m := <-s.measurements:
			log.Printf("Received measurement from sensor %v", m.Name)
			for _, e := range s.Exporters {
				log.Printf("Exporting measurement to %v", e.Name())
				if err := e.Export(ctx, m); err != nil {
					log.Printf("Failed to report measurement: %v", err)
				}
			}
		case <-s.quit:
			return
		}
	}
}
