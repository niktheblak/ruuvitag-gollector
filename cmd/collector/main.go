package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/config"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

func run(c *cli.Context) error {
	log.Println("Starting ruuvitag-gollector")
	log.Printf("Device: %v", c.GlobalString("device"))
	cfg, err := config.Read(c.GlobalString("config"))
	if err != nil {
		return fmt.Errorf("failed to decode configuration: %w", err)
	}
	err = cfg.Validate()
	if err != nil {
		return err
	}
	ruuviTags, err := parseRuuviTags(c.GlobalStringSlice("ruuvitags"))
	if err != nil {
		return err
	}
	scn, err := scanner.New(c.GlobalDuration("reporting_interval"), c.String("device"), ruuviTags)
	if err != nil {
		return fmt.Errorf("failed to create scanner: %w", err)
	}
	var exporters []exporter.Exporter
	if c.GlobalBool("console") {
		exporters = append(exporters, console.Exporter{})
	}
	ctx := context.Background()
	if c.GlobalBool("influxdb") {
		influx, err := influxdb.New(influxdb.Config{
			URL:      c.GlobalString("url"),
			Username: c.GlobalString("username"),
			Password: c.GlobalString("password"),
		})
		if err != nil {
			return fmt.Errorf("failed to create InfluxDB reporter: %w", err)
		}
		exporters = append(exporters, influx)
	}
	if c.GlobalBool("pubsub") {
		ps, err := pubsub.New(ctx, c.GlobalString("project"), c.GlobalString("topic"))
		if err != nil {
			return fmt.Errorf("failed to create Google Pub/Sub reporter: %w", err)
		}
		exporters = append(exporters, ps)
	}
	scn.Exporters = exporters
	log.Println("Starting scanner")
	if err := scn.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scanner: %w", err)
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
	return nil
}

func parseRuuviTags(ruuviTags []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, rt := range ruuviTags {
		tokens := strings.SplitN(rt, "=", 2)
		if len(tokens) != 2 {
			return nil, fmt.Errorf("invalid RuuviTag entry: %s", rt)
		}
		key := strings.TrimSpace(tokens[0])
		value := strings.TrimSpace(tokens[1])
		m[key] = value
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
			Name:     "config",
			Usage:    "RuuviTag configuration file",
			Required: true,
		},
		cli.BoolFlag{
			Name:  "daemon, s",
			Usage: "run as a background service",
		},
		cli.BoolFlag{
			Name:  "console, c",
			Usage: "print measurements to console",
		},
		altsrc.NewDurationFlag(cli.DurationFlag{
			Name:  "reporting_interval, r",
			Usage: "reporting interval",
			Value: 60 * time.Second,
		}),
		altsrc.NewStringSliceFlag(cli.StringSliceFlag{
			Name:  "ruuvitags",
			Usage: "RuuviTag addresses and names to use",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "device",
			Usage: "HCL device to use",
			Value: "default",
		}),
		cli.BoolFlag{
			Name:   "influxdb",
			Usage:  "use influxdb",
			EnvVar: "RUUVITAG_USE_INFLUXDB",
		},
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "url",
			Usage:  "InfluxDB URL",
			EnvVar: "RUUVITAG_INFLUXDB_URL",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "username, u",
			Usage:  "InfluxDB username",
			EnvVar: "RUUVITAG_INFLUXDB_USERNAME",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "password, p",
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
			Usage:  "Google Pub/Sub project",
			EnvVar: "RUUVITAG_PUBSUB_PROJECT",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "topic",
			Usage:  "Google Pub/Sub topic",
			EnvVar: "RUUVITAG_PUBSUB_TOPIC",
		}),
	}
	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("config"))
	app.Action = func(c *cli.Context) error {
		return run(c)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
