package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/config"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var sleepInterval = 60 * time.Second
var quit = make(chan int)
var stopScan = make(chan int)
var measurements = make(chan ruuvitag.SensorData, 10)
var ruuviTags []gatt.UUID
var exporters []exporter.Exporter

func beginScan(d gatt.Device) {
	timer := time.NewTimer(sleepInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			d.Scan(ruuviTags, false)
		case <-stopScan:
			return
		case <-quit:
			return
		}
	}
}

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		go beginScan(d)
	case gatt.StatePoweredOff:
		stopScan <- 1
		if err := d.Init(onStateChanged); err != nil {
			log.Fatalf("Failed to initialize device: %v", err)
		}
	default:
		log.Printf("Unhandled state: %s", s)
	}
}

func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	data, err := ruuvitag.Parse(a.ManufacturerData)
	if err != nil {
		log.Printf("Error while parsing RuuviTag data: %v", err)
		return
	}
	data.DeviceID = p.ID()
	data.Timestamp = time.Now()
	log.Printf("Read sensor data %v from device ID %v", data, p.ID())
	measurements <- data
}

func exportMeasurements(ctx context.Context) {
	for {
		select {
		case m := <-measurements:
			log.Printf("Received measurement from sensor %v", m.Name)
			for _, e := range exporters {
				log.Printf("Exporting measurement to %v", e.Name())
				if err := e.Export(ctx, m); err != nil {
					log.Printf("Failed to report measurement: %v", err)
				}
			}
		case <-quit:
			return
		}
	}
}

func initRuuviTags(cfg config.Config) {
	for _, rt := range cfg.RuuviTag {
		// TODO: convert MACs to UUIDs?
		uid, err := gatt.ParseUUID(rt.MAC)
		if err != nil {
			log.Fatalf("Failed to parse RuuviTag UUID %s: %v", rt.MAC, err)
		}
		ruuviTags = append(ruuviTags, uid)
	}
	log.Printf("Reading from RuuviTags %v", ruuviTags)
}

func initInfluxdbExporter() {
	influxEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_INFLUXDB"))
	if influxEnabled {
		influx, err := influxdb.New()
		if err != nil {
			log.Fatalf("Failed to create InfluxDB reporter: %v", err)
		}
		exporters = append(exporters, influx)
	}
}

func initGooglePubsubExporter() {
	pubsubEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_PUBSUB"))
	if pubsubEnabled {
		ps, err := pubsub.New()
		if err != nil {
			log.Fatalf("Failed to create Google Pub/Sub reporter: %v", err)
		}
		exporters = append(exporters, ps)
	}
}

func main() {
	cfg, err := config.ReadConfig("ruuvitags.toml")
	if err != nil {
		log.Fatalf("Failed to decode configuration: %v", err)
	}
	err = cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}
	sleepInterval = cfg.ReportingInterval.Duration
	initRuuviTags(cfg)
	exporters = append(exporters, console.Exporter{})
	initInfluxdbExporter()
	initGooglePubsubExporter()
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device: %v", err)
	}
	log.Println("Starting ruuvitag-gollector")
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	if err := device.Init(onStateChanged); err != nil {
		log.Fatalf("Failed to initialize device: %v", err)
	}
	ctx := context.Background()
	go exportMeasurements(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Stopping ruuvitag-gollector")
	for _, e := range exporters {
		e.Close()
	}
	quit <- 1
}
