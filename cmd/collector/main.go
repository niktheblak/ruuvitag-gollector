package main

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/reporter/console"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/reporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/reporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

var sleepInterval = 60 * time.Second
var quit = make(chan int)
var stopScan = make(chan int)
var measurements = make(chan ruuvitag.SensorData, 10)
var reporters []reporter.Reporter

func beginScan(d gatt.Device) {
	timer := time.NewTimer(sleepInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			d.Scan(nil, true)
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

func reportMeasurements() {
	for {
		select {
		case m := <-measurements:
			log.Printf("Received measurement from sensor %v", m.Name)
			for _, r := range reporters {
				log.Printf("Reporting measurement to %v", r.Name())
				if err := r.Report(m); err != nil {
					log.Printf("Failed to report measurement: %v", err)
				}
			}
		case <-quit:
			return
		}
	}
}

func main() {
	d, err := time.ParseDuration(os.Getenv("RUUVITAG_REPORTING_INTERVAL"))
	if err == nil {
		sleepInterval = d
	}
	reporters = append(reporters, console.Reporter{})
	influxEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_INFLUXDB"))
	if influxEnabled {
		influx, err := influxdb.New()
		if err != nil {
			log.Fatalf("Failed to create InfluxDB reporter: %v", err)
		}
		reporters = append(reporters, influx)
	}
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device: %v", err)
	}
	log.Println("Starting ruuvitag-gollector")
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	if err := device.Init(onStateChanged); err != nil {
		log.Fatalf("Failed to initialize device: %v", err)
	}
	go reportMeasurements()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Stopping ruuvitag-gollector")
	quit <- 1
}
