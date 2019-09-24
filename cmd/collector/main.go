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
var ruuviTagDeviceIDs []gatt.UUID
var ruuviTagNames map[string]string
var exporters []exporter.Exporter

func beginScan(d gatt.Device) {
	timer := time.NewTimer(sleepInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			d.Scan(ruuviTagDeviceIDs, false)
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
	data.Name = ruuviTagNames[p.ID()]
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
	ruuviTagNames = make(map[string]string)
	for _, rt := range cfg.RuuviTags {
		ruuviTagNames[rt.ID] = rt.Name
		uid, err := gatt.ParseUUID(rt.ID)
		if err != nil {
			log.Fatalf("Failed to parse RuuviTag UUID %s: %v", rt.ID, err)
		}
		ruuviTagDeviceIDs = append(ruuviTagDeviceIDs, uid)
	}
	log.Printf("Reading from RuuviTags %v", ruuviTagDeviceIDs)
}

func initInfluxdbExporter() {
	influxEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_INFLUXDB"))
	if !influxEnabled {
		return
	}
	url := os.Getenv("RUUVITAG_INFLUXDB_URL")
	if url == "" {
		log.Fatal("RUUVITAG_INFLUXDB_URL must be set")
	}
	influx, err := influxdb.New(influxdb.Config{
		URL:      url,
		Username: os.Getenv("RUUVITAG_INFLUXDB_USERNAME"),
		Password: os.Getenv("RUUVITAG_INFLUXDB_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("Failed to create InfluxDB reporter: %v", err)
	}
	exporters = append(exporters, influx)
}

func initGooglePubsubExporter(ctx context.Context) {
	pubsubEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_PUBSUB"))
	if !pubsubEnabled {
		return
	}
	project := os.Getenv("RUUVITAG_PUBSUB_PROJECT")
	if project == "" {
		log.Fatal("RUUVITAG_PUBSUB_PROJECT must be set")
	}
	topic := os.Getenv("RUUVITAG_PUBSUB_TOPIC")
	if topic == "" {
		log.Fatal("RUUVITAG_PUBSUB_TOPIC must be set")
	}
	ps, err := pubsub.New(ctx, project, topic)
	if err != nil {
		log.Fatalf("Failed to create Google Pub/Sub reporter: %v", err)
	}
	exporters = append(exporters, ps)
}

func main() {
	configFile := os.Getenv("RUUVITAG_CONFIG_FILE")
	if configFile == "" {
		configFile = "ruuvitags.toml"
	}
	cfg, err := config.ReadConfig(configFile)
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
	ctx := context.Background()
	initInfluxdbExporter()
	initGooglePubsubExporter(ctx)
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device: %v", err)
	}
	log.Println("Starting ruuvitag-gollector")
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	if err := device.Init(onStateChanged); err != nil {
		log.Fatalf("Failed to initialize device: %v", err)
	}
	go exportMeasurements(ctx)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Stopping ruuvitag-gollector")
	for _, e := range exporters {
		if err := e.Close(); err != nil {
			log.Println(err)
		}
	}
	quit <- 1
}
