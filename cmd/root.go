package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-ble/ble"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/http"
)

var ErrNotEnabled = errors.New("this exporter is not included in the build")

var (
	logger      *zap.Logger
	peripherals map[string]string
	exporters   []exporter.Exporter
	cfgFile     string
	device      string
)

var rootCmd = &cobra.Command{
	Use:               "ruuvitag-gollector",
	Short:             "Collects measurements from RuuviTag sensors",
	SilenceUsage:      true,
	PersistentPreRunE: run,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if logger != nil {
			logger.Sync()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ruuvitag-gollector.yaml)")

	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().String("device", "default", "HCL device to use")
	rootCmd.PersistentFlags().BoolP("console", "c", false, "Print measurements to console")
	rootCmd.PersistentFlags().String("loglevel", "info", "Log level")

	rootCmd.PersistentFlags().Bool("http.enabled", false, "Send measurements as JSON to a HTTP endpoint")
	rootCmd.PersistentFlags().String("http.addr", "", "HTTP receiver address")
	rootCmd.PersistentFlags().String("http.token", "", "HTTP receiver authorization token")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigName("ruuvitag-gollector")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	creds := viper.GetString("gcp.credentials")
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return err
		}
	}
	if viper.GetBool("gcp.stackdriver.enabled") {
		if err := initStackdriverLogging(); err != nil {
			return fmt.Errorf("failed to create Stackdriver logger: %w", err)
		}
	} else {
		logLevel := viper.GetString("loglevel")
		if logLevel == "" {
			logLevel = "info"
		}
		var zapLogLevel zap.AtomicLevel
		if err := zapLogLevel.UnmarshalText([]byte(logLevel)); err != nil {
			return err
		}
		cfg := zap.Config{
			Level:            zapLogLevel,
			DisableCaller:    true,
			Encoding:         "console",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:          zapcore.OmitKey,
				LevelKey:         "L",
				NameKey:          zapcore.OmitKey,
				CallerKey:        zapcore.OmitKey,
				FunctionKey:      zapcore.OmitKey,
				MessageKey:       "M",
				StacktraceKey:    "S",
				LineEnding:       zapcore.DefaultLineEnding,
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				EncodeDuration:   zapcore.StringDurationEncoder,
				ConsoleSeparator: " ",
			},
		}
		var err error
		logger, err = cfg.Build()
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}
	}
	ruuviTags := viper.GetStringMapString("ruuvitags")
	logger.Info("RuuviTags", zap.Any("ruuvitags", ruuviTags))
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
