package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func run(c *cli.Context) error {
	ruuviTags, err := parseRuuviTags(c.GlobalStringSlice("ruuvitags"))
	if err != nil {
		return err
	}
	if c.GlobalBool("stackdriver") {
		project := c.GlobalString("project")
		if project == "" {
			return fmt.Errorf("Google Cloud Platform project must be specified")
		}
		ctx := context.Background()
		client, err := logging.NewClient(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to create Stackdriver client: %w", err)
		}
		defer client.Close()
		logger = client.Logger("ruuvitag-gollector").StandardLogger(logging.Info)
	}
	scn, err := scanner.New(logger, c.String("device"), ruuviTags)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}
	defer scn.Close()
	var exporters []exporter.Exporter
	if c.GlobalBool("console") {
		exporters = append(exporters, console.Exporter{})
	}
	if c.GlobalBool("influxdb") {
		url := c.GlobalString("influxdb_addr")
		if url == "" {
			return fmt.Errorf("InfluxDB address must be specified")
		}
		influx, err := influxdb.New(influxdb.Config{
			Addr:        url,
			Database:    c.GlobalString("influxdb_database"),
			Measurement: c.GlobalString("influxdb_measurement"),
			Username:    c.GlobalString("influxdb_username"),
			Password:    c.GlobalString("influxdb_password"),
		})
		if err != nil {
			return fmt.Errorf("failed to create InfluxDB reporter: %w", err)
		}
		exporters = append(exporters, influx)
	}
	if c.GlobalBool("pubsub") {
		ctx := context.Background()
		project := c.GlobalString("project")
		if project == "" {
			return fmt.Errorf("Google Cloud Platform project must be specified")
		}
		topic := c.GlobalString("pubsub_topic")
		if topic == "" {
			return fmt.Errorf("Google Pub/Sub topic must be specified")
		}
		ps, err := pubsub.New(ctx, project, topic)
		if err != nil {
			return fmt.Errorf("failed to create Google Pub/Sub reporter: %w", err)
		}
		exporters = append(exporters, ps)
	}
	scn.Exporters = exporters
	logger.Println("Starting ruuvitag-gollector")
	if c.GlobalBool("daemon") {
		return runAsDaemon(scn, c.GlobalDuration("scan_interval"))
	} else {
		return runOnce(scn)
	}
}

func runAsDaemon(scn *scanner.Scanner, scanInterval time.Duration) error {
	logger.Println("Starting scanner")
	ctx := context.Background()
	var err error
	if scanInterval > 0 {
		err = scn.ScanWithInterval(ctx, scanInterval)
	} else {
		err = scn.ScanContinuously(ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to start scanner: %w", err)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	logger.Println("Stopping ruuvitag-gollector")
	scn.Stop()
	return nil
}

func runOnce(scn *scanner.Scanner) error {
	logger.Println("Scanning once")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		cancel()
		scn.Stop()
	}()
	if err := scn.ScanOnce(ctx); err != nil {
		logger.Printf("failed to scan: %v", err)
	}
	logger.Println("Stopping ruuvitag-gollector")
	return nil
}

func parseRuuviTags(ruuviTags []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, rt := range ruuviTags {
		tokens := strings.SplitN(rt, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("invalid RuuviTag entry: %s", rt)
		}
		addr := strings.ToLower(strings.TrimSpace(tokens[0]))
		name := strings.TrimSpace(tokens[1])
		m[addr] = name
	}
	return m, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "ruuvitag-gollector"
	app.Usage = "Collect measurements from RuuviTag sensors"
	app.Version = "1.0.0"
	app.Author = "Niko Korhonen"
	app.Email = "niko@bitnik.fi"
	app.Copyright = "(c) 2019 Niko Korhonen"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:      "config",
			Usage:     "RuuviTag configuration file",
			EnvVar:    "RUUVITAG_CONFIG_FILE",
			TakesFile: true,
			Required:  true,
		},
		cli.BoolFlag{
			Name:  "daemon, d",
			Usage: "run as a background service",
		},
		cli.BoolFlag{
			Name:  "console, c",
			Usage: "print measurements to console",
		},
		altsrc.NewStringSliceFlag(cli.StringSliceFlag{
			Name:  "ruuvitags",
			Usage: "RuuviTag addresses and names to use",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "device",
			Usage: "HCL device to use",
			Value: "default",
		}),
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:   "scan_interval",
			Usage:  "Pause between RuuviTag device scans in daemon mode",
			EnvVar: "RUUVITAG_SCAN_INTERVAL",
			Value:  1 * time.Minute,
		}),
		cli.BoolFlag{
			Name:   "influxdb",
			Usage:  "use influxdb",
			EnvVar: "RUUVITAG_USE_INFLUXDB",
		},
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "influxdb_addr",
			Usage:  "InfluxDB server address",
			EnvVar: "RUUVITAG_INFLUXDB_ADDR",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "influxdb_database",
			Usage:  "InfluxDB database",
			EnvVar: "RUUVITAG_INFLUXDB_DATABASE",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "influxdb_measurement",
			Usage:  "InfluxDB measurement",
			EnvVar: "RUUVITAG_INFLUXDB_MEASUREMENT",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "influxdb_username",
			Usage:  "InfluxDB username",
			EnvVar: "RUUVITAG_INFLUXDB_USERNAME",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "influxdb_password",
			Usage:  "InfluxDB password",
			EnvVar: "RUUVITAG_INFLUXDB_PASSWORD",
		}),
		cli.BoolFlag{
			Name:   "pubsub",
			Usage:  "use Google Pub/Sub",
			EnvVar: "RUUVITAG_USE_PUBSUB",
		},
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "project",
			Usage:  "Google Cloud Platform project",
			EnvVar: "RUUVITAG_GOOGLE_PROJECT",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "pubsub_topic",
			Usage:  "Google Pub/Sub topic",
			EnvVar: "RUUVITAG_PUBSUB_TOPIC",
		}),
		cli.BoolFlag{
			Name:   "stackdriver",
			Usage:  "use Google Stackdriver logging",
			EnvVar: "RUUVITAG_USE_STACKDRIVER_LOGGING",
		},
	}
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("config"))
	app.Action = func(c *cli.Context) error {
		return run(c)
	}
	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
