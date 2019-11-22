package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/niktheblak/gcloudzap"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	logger  *zap.Logger
	scn     *scanner.Scanner
	cfgFile string
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
		if scn != nil {
			scn.Close()
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ruuvitag-gollector.yaml)")

	rootCmd.PersistentFlags().StringToString("ruuvitags", nil, "RuuviTag addresses and names to use")
	rootCmd.PersistentFlags().String("device", "default", "HCL device to use")
	rootCmd.PersistentFlags().BoolP("console", "c", false, "Print measurements to console")

	rootCmd.PersistentFlags().Bool("influxdb", false, "Store measurements to InfluxDB")
	rootCmd.PersistentFlags().String("influxdb_addr", "", "InfluxDB address with protocol, host and port")
	rootCmd.PersistentFlags().String("influxdb_database", "", "InfluxDB database to use ")
	rootCmd.PersistentFlags().String("influxdb_measurement", "", "InfluxDB measurement name")
	rootCmd.PersistentFlags().String("influxdb_username", "", "InfluxDB username")
	rootCmd.PersistentFlags().String("influxdb_password", "", "InfluxDB password")

	rootCmd.PersistentFlags().Bool("gcp_stackdriver", false, "Send logs to Google Stackdriver")
	rootCmd.PersistentFlags().String("gcp_credentials", "", "Google Cloud application credentials file")
	rootCmd.MarkFlagFilename("gcp_credentials", "json")
	rootCmd.PersistentFlags().String("gcp_project", "", "Google Cloud Platform project")
	rootCmd.PersistentFlags().Bool("gcp_pubsub", false, "Send measurements to Google Pub/Sub")
	rootCmd.PersistentFlags().String("gcp_pubsub_topic", "", "Google Pub/Sub topic to use")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".ruuvitag-gollector")
	}

	viper.SetEnvPrefix("ruuvitag")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func run(cmd *cobra.Command, args []string) error {
	creds := viper.GetString("gcp_credentials")
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return err
		}
	}
	if viper.GetBool("gcp_stackdriver") {
		project := viper.GetString("gcp_project")
		if project == "" {
			return fmt.Errorf("Google Cloud Platform project must be specified")
		}
		var err error
		logger, err = gcloudzap.NewProduction(project, "ruuvitag-gollector")
		if err != nil {
			return fmt.Errorf("failed to create Stackdriver logger: %w", err)
		}
	} else {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}
	}
	ruuviTags := viper.GetStringMapString("ruuvitags")
	logger.Info("RuuviTags", zap.Any("ruuvitags", ruuviTags))
	scn = scanner.New(logger, ruuviTags)
	var exporters []exporter.Exporter
	if viper.GetBool("console") {
		exporters = append(exporters, console.Exporter{})
	}
	if viper.GetBool("influxdb") {
		addr := viper.GetString("influxdb_addr")
		if addr == "" {
			return fmt.Errorf("InfluxDB address must be specified")
		}
		influx, err := influxdb.New(influxdb.Config{
			Addr:        addr,
			Database:    viper.GetString("influxdb_database"),
			Measurement: viper.GetString("influxdb_measurement"),
			Username:    viper.GetString("influxdb_username"),
			Password:    viper.GetString("influxdb_password"),
		})
		if err != nil {
			return fmt.Errorf("failed to create InfluxDB reporter: %w", err)
		}
		exporters = append(exporters, influx)
	}
	if viper.GetBool("gcp_pubsub") {
		ctx := context.Background()
		project := viper.GetString("gcp_project")
		if project == "" {
			return fmt.Errorf("Google Cloud Platform project must be specified")
		}
		topic := viper.GetString("gcp_pubsub_topic")
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
	device := viper.GetString("device")
	logger.Info("Initializing new device", zap.String("device", device))
	if err := scn.Init(device); err != nil {
		return err
	}
	return nil
}
