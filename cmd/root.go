package cmd

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/go-ble/ble"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/http"
)

var ErrNotEnabled = errors.New("this exporter is not included in the build")

var (
	logger      *slog.Logger
	peripherals map[string]string
	exporters   []exporter.Exporter
	device      string
)

var rootCmd = &cobra.Command{
	Use:               "ruuvitag-gollector",
	Short:             "Collects measurements from RuuviTag sensors",
	SilenceUsage:      true,
	PersistentPreRunE: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	logger = slog.Default()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().String("device", "default", "HCL device to use")
	rootCmd.PersistentFlags().BoolP("console", "c", false, "Print measurements to console")
	rootCmd.PersistentFlags().String("loglevel", "info", "Log level")

	rootCmd.PersistentFlags().Bool("http.enabled", false, "Send measurements as JSON to a HTTP endpoint")
	rootCmd.PersistentFlags().String("http.addr", "", "HTTP receiver address")
	rootCmd.PersistentFlags().String("http.token", "", "HTTP receiver authorization token")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}
}

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/ruuvitag-gollector/")
	viper.AddConfigPath("$HOME/.ruuvitag-gollector")
	if err := viper.ReadInConfig(); err != nil {
		logger.LogAttrs(nil, slog.LevelInfo, "Config file does not exist, using only command line arguments", slog.String("file", viper.ConfigFileUsed()))
	} else {
		logger.LogAttrs(nil, slog.LevelInfo, "Read config from file", slog.String("file", viper.ConfigFileUsed()))
	}
}

func run(_ *cobra.Command, _ []string) error {
	creds := viper.GetString("gcp.credentials")
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return err
		}
	}
	logLevel := viper.GetString("loglevel")
	if logLevel == "" {
		logLevel = "info"
	}
	var programLevel = new(slog.LevelVar)
	if err := programLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return err
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	logger = slog.New(h)
	ruuviTags := viper.GetStringMapString("ruuvitags")
	if len(ruuviTags) == 0 {
		logger.LogAttrs(nil, slog.LevelError, "At least one RuuviTag address must be specified")
		os.Exit(1)
	}
	logger.LogAttrs(nil, slog.LevelInfo, "RuuviTags", slog.Any("ruuvitags", ruuviTags))
	peripherals = make(map[string]string)
	for addr, name := range ruuviTags {
		peripherals[ble.NewAddr(addr).String()] = name
	}
	if viper.GetBool("console") {
		exporters = append(exporters, console.Exporter{})
	}
	if viper.GetBool("influxdb.enabled") {
		if err := addInfluxDBExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create InfluxDB exporter: %w", err)
		}
	}
	if viper.GetBool("gcp.pubsub.enabled") {
		if err := addPubSubExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create Google Pub/Sub exporter: %w", err)
		}
	}
	if viper.GetBool("aws.dynamodb.enabled") {
		if err := addDynamoDBExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create AWS DynamoDB exporter: %w", err)
		}
	}
	if viper.GetBool("aws.sqs.enabled") {
		if err := addSQSExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create AWS SQS exporter: %w", err)
		}
	}
	if viper.GetBool("postgres.enabled") {
		if err := addPostgresExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create PostgreSQL exporter: %w", err)
		}
	}
	if viper.GetBool("http.enabled") {
		addr := viper.GetString("http.addr")
		token := viper.GetString("http.token")
		exp, err := http.New(addr, token, 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to create HTTP exporter: %w", err)
		}
		exporters = append(exporters, exp)
	}
	if viper.GetBool("mqtt.enabled") {
		if err := addMQTTExporter(&exporters); err != nil {
			return fmt.Errorf("failed to create MQTT exporter: %w", err)
		}
	}
	device = viper.GetString("device")
	return nil
}
