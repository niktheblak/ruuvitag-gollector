package scanner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/config"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

type Scanner struct {
	SleepInterval     time.Duration
	Exporters         []exporter.Exporter
	quit              chan int
	stopScan          chan int
	measurements      chan ruuvitag.SensorData
	ruuviTagDeviceIDs []gatt.UUID
	ruuviTagNames     map[string]string
}

func New(cfg config.Config) (*Scanner, error) {
	scn := &Scanner{
		SleepInterval: cfg.ReportingInterval.Duration,
		quit:          make(chan int, 1),
		stopScan:      make(chan int, 1),
		measurements:  make(chan ruuvitag.SensorData, 10),
		ruuviTagNames: make(map[string]string),
	}
	for _, rt := range cfg.RuuviTags {
		scn.ruuviTagNames[rt.ID] = rt.Name
		uid, err := gatt.ParseUUID(rt.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RuuviTag UUID %s: %w", rt.ID, err)
		}
		scn.ruuviTagDeviceIDs = append(scn.ruuviTagDeviceIDs, uid)
	}
	if len(scn.ruuviTagDeviceIDs) > 0 {
		log.Printf("Reading from RuuviTags %v", scn.ruuviTagDeviceIDs)
	} else {
		log.Println("Reading from all nearby BLE devices")
	}
	return scn, nil
}

func (s *Scanner) Start(ctx context.Context) error {
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		return fmt.Errorf("failed to open device: %w", err)
	}
	device.Handle(gatt.PeripheralDiscovered(s.onPeripheralDiscovered))
	if err := device.Init(s.onStateChanged); err != nil {
		return fmt.Errorf("failed to initialize device: %w", err)
	}
	go s.exportMeasurements(ctx)
	return nil
}

func (s *Scanner) Stop() {
	s.quit <- 1
}

func (s *Scanner) beginScan(d gatt.Device) {
	log.Println("Scanner starting")
	timer := time.NewTimer(s.SleepInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			log.Printf("Scanner scanning devices %v", s.ruuviTagDeviceIDs)
			d.Scan(s.ruuviTagDeviceIDs, false)
		case <-s.stopScan:
			log.Println("Scanner stopping")
			return
		case <-s.quit:
			log.Println("Scanner quitting")
			return
		}
	}
}

func (s *Scanner) onStateChanged(d gatt.Device, state gatt.State) {
	switch state {
	case gatt.StatePoweredOn:
		log.Println("Device powered on")
		go s.beginScan(d)
	case gatt.StatePoweredOff:
		log.Println("Device powered off")
		s.stopScan <- 1
		// Attempt to restart device
		if err := d.Init(s.onStateChanged); err != nil {
			log.Printf("Failed to restart device: %v", err)
		}
	default:
		log.Printf("Unhandled state: %v", state)
	}
}

func (s *Scanner) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	data, err := ruuvitag.Parse(a.ManufacturerData)
	if err != nil {
		log.Printf("Error while parsing RuuviTag data: %v", err)
		return
	}
	data.DeviceID = p.ID()
	data.Name = s.ruuviTagNames[p.ID()]
	data.Timestamp = time.Now()
	log.Printf("Read sensor data %v from device ID %v", data, p.ID())
	s.measurements <- data
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
