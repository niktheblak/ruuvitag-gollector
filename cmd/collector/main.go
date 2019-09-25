package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/niktheblak/ruuvitag-gollector/pkg/config"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

func createInfluxdbExporter() exporter.Exporter {
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
	return influx
}

func createGooglePubsubExporter(ctx context.Context) exporter.Exporter {
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
	return ps
}

func main() {
	log.Println("Starting ruuvitag-gollector")
	configFile := os.Getenv("RUUVITAG_CONFIG_FILE")
	if configFile == "" {
		configFile = "ruuvitags.toml"
	}
	cfg, err := config.Read(configFile)
	if err != nil {
		log.Fatalf("Failed to decode configuration: %v", err)
	}
	err = cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}
	scn, err := scanner.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create scanner: %v", err)
	}
	var exporters []exporter.Exporter
	exporters = append(exporters, console.Exporter{})
	ctx := context.Background()
	influxEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_INFLUXDB"))
	if influxEnabled {
		exporters = append(exporters, createInfluxdbExporter())
	}
	pubsubEnabled, _ := strconv.ParseBool(os.Getenv("RUUVITAG_USE_PUBSUB"))
	if pubsubEnabled {
		exporters = append(exporters, createGooglePubsubExporter(ctx))
	}
	scn.Exporters = exporters
	log.Println("Starting scanner")
	if err := scn.Start(ctx); err != nil {
		log.Fatalf("Failed to start scanner: %v", err)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Stopping ruuvitag-gollector")
	scn.Stop()
	for _, e := range exporters {
		if err := e.Close(); err != nil {
			log.Printf("Failed to close exporter %s: %v", e.Name(), err)
		}
	}
}
