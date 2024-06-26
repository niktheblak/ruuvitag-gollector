package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
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
	cfgFile     string
	logger      *slog.Logger
	peripherals map[string]string
	exporters   []exporter.Exporter
	device      string
)

var rootCmd = &cobra.Command{
	Use:               "ruuvitag-gollector",
	Short:             "Collects measurements from RuuviTag sensors",
	SilenceUsage:      true,
	PersistentPreRunE: preRun,
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Shutting down")
		for _, exp := range exporters {
			if err := exp.Close(); err != nil {
				return err
			}
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	logger = slog.Default()

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().String("device", "", "HCL device to use")
	rootCmd.PersistentFlags().BoolP("console", "c", false, "Print measurements to console")
	rootCmd.PersistentFlags().String("loglevel", "info", "Log level")
	rootCmd.PersistentFlags().String("log.level", "info", "Log level")
	rootCmd.PersistentFlags().String("log.format", "text", "Log level")

	rootCmd.PersistentFlags().Bool("http.enabled", false, "Send measurements as JSON to a HTTP endpoint")
	rootCmd.PersistentFlags().String("http.addr", "", "HTTP receiver address")
	rootCmd.PersistentFlags().String("http.token", "", "HTTP receiver authorization token")

	viper.SetDefault("loglevel", "info")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("device", "default")
}

func initConfig() {
	cobra.CheckErr(viper.BindPFlags(rootCmd.PersistentFlags()))
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.ruuvitag-gollector")
		viper.AddConfigPath("/etc/ruuvitag-gollector/")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	configErr := viper.ReadInConfig()
	logLevelCfg := viper.GetString("log.level")
	if logLevelCfg == "" {
		logLevelCfg = viper.GetString("loglevel")
	}
	var logLevel = new(slog.LevelVar)
	if err := logLevel.UnmarshalText([]byte(logLevelCfg)); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid log level: %s\n", err)
		os.Exit(1)
	}
	logFormat := viper.GetString("log.format")
	var logHandler slog.Handler
	switch logFormat {
	case "text":
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	default:
		fmt.Fprintf(os.Stderr, "Invalid log format: %s\n", logFormat)
		os.Exit(1)
	}
	logger = slog.New(logHandler)
	if configErr != nil {
		logger.LogAttrs(nil, slog.LevelInfo, "Config file not found, using only command line arguments", slog.String("file", viper.ConfigFileUsed()))
	} else {
		logger.LogAttrs(nil, slog.LevelInfo, "Read config from file", slog.String("file", viper.ConfigFileUsed()))
	}
}

func preRun(_ *cobra.Command, _ []string) error {
	creds := viper.GetString("gcp.credentials")
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return err
		}
	}
	ruuviTags := viper.GetStringMapString("ruuvitags")
	if len(ruuviTags) == 0 {
		return fmt.Errorf("at least one RuuviTag address must be specified")
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
