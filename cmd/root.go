package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/niktheblak/gcloudzap"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/dynamodb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/sqs"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/gcp/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	logger    *zap.Logger
	scn       *scanner.Scanner
	ruuviTags map[string]string
	cfgFile   string
	device    string
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

	rootCmd.PersistentFlags().Bool("influxdb.enabled", false, "Store measurements to InfluxDB")
	rootCmd.PersistentFlags().String("influxdb.addr", "http://localhost:8086", "InfluxDB address with protocol, host and port")
	rootCmd.PersistentFlags().String("influxdb.database", "", "InfluxDB database to use")
	rootCmd.PersistentFlags().String("influxdb.measurement", "", "InfluxDB measurement name")
	rootCmd.PersistentFlags().String("influxdb.username", "", "InfluxDB username")
	rootCmd.PersistentFlags().String("influxdb.password", "", "InfluxDB password")

	rootCmd.PersistentFlags().Bool("gcp.stackdriver.enabled", false, "Send logs to Google Stackdriver")
	rootCmd.PersistentFlags().String("gcp.credentials", "", "Google Cloud application credentials file")
	rootCmd.MarkFlagFilename("gcp.credentials", "json")
	rootCmd.PersistentFlags().String("gcp.project", "", "Google Cloud Platform project")
	rootCmd.PersistentFlags().Bool("gcp.pubsub.enabled", false, "Send measurements to Google Pub/Sub")
	rootCmd.PersistentFlags().String("gcp.pubsub.topic", "", "Google Pub/Sub topic to use")

	rootCmd.PersistentFlags().String("aws.region", "us-east-2", "AWS region")
	rootCmd.PersistentFlags().String("aws.access_key_id", "", "AWS access key ID")
	rootCmd.PersistentFlags().String("aws.secret_access_key", "", "AWS secret access key")
	rootCmd.PersistentFlags().String("aws.session_token", "", "AWS session token")
	rootCmd.PersistentFlags().Bool("aws.dynamodb.enabled", false, "Store measurements to AWS DynamoDB")
	rootCmd.PersistentFlags().String("aws.dynamodb.table", "", "AWS DynamoDB table name")
	rootCmd.PersistentFlags().Bool("aws.sqs.enabled", false, "Send measurements to AWS SQS")
	rootCmd.PersistentFlags().String("aws.sqs.queue", "", "AWS SQS queue name")

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
	creds := viper.GetString("gcp.credentials")
	if creds != "" {
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", creds); err != nil {
			return err
		}
	}
	if viper.GetBool("gcp.stackdriver.enabled") {
		project := viper.GetString("gcp.project")
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
	ruuviTags = viper.GetStringMapString("ruuvitags")
	logger.Info("RuuviTags", zap.Any("ruuvitags", ruuviTags))
	scn = scanner.New(logger, ruuviTags)
	var exporters []exporter.Exporter
	if viper.GetBool("console") {
		exporters = append(exporters, console.Exporter{})
	}
	if viper.GetBool("influxdb.enabled") {
		addr := viper.GetString("influxdb.addr")
		if addr == "" {
			return fmt.Errorf("InfluxDB address must be specified")
		}
		influx, err := influxdb.New(influxdb.Config{
			Addr:        addr,
			Database:    viper.GetString("influxdb.database"),
			Measurement: viper.GetString("influxdb.measurement"),
			Username:    viper.GetString("influxdb.username"),
			Password:    viper.GetString("influxdb.password"),
		})
		if err != nil {
			return fmt.Errorf("failed to create InfluxDB exporter: %w", err)
		}
		exporters = append(exporters, influx)
	}
	if viper.GetBool("gcp.pubsub.enabled") {
		ctx := context.Background()
		project := viper.GetString("gcp.project")
		if project == "" {
			return fmt.Errorf("Google Cloud Platform project must be specified")
		}
		topic := viper.GetString("gcp.pubsub.topic")
		if topic == "" {
			return fmt.Errorf("Google Pub/Sub topic must be specified")
		}
		ps, err := pubsub.New(ctx, project, topic)
		if err != nil {
			return fmt.Errorf("failed to create Google Pub/Sub exporter: %w", err)
		}
		exporters = append(exporters, ps)
	}
	if viper.GetBool("aws.dynamodb.enabled") {
		exp, err := dynamodb.New(dynamodb.Config{
			Table:           viper.GetString("aws.dynamodb.table"),
			Region:          viper.GetString("aws.region"),
			AccessKeyID:     viper.GetString("aws.access_key_id"),
			SecretAccessKey: viper.GetString("aws.secret_access_key"),
			SessionToken:    viper.GetString("aws.session_token"),
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS DynamoDB exporter: %w", err)
		}
		exporters = append(exporters, exp)
	}
	if viper.GetBool("aws.sqs.enabled") {
		exp, err := sqs.New(sqs.Config{
			Queue:           viper.GetString("aws.sqs.queue"),
			Region:          viper.GetString("aws.region"),
			AccessKeyID:     viper.GetString("aws.access_key_id"),
			SecretAccessKey: viper.GetString("aws.secret_access_key"),
			SessionToken:    viper.GetString("aws.session_token"),
		})
		if err != nil {
			return fmt.Errorf("failed to create AWS SQS exporter: %w", err)
		}
		exporters = append(exporters, exp)
	}
	scn.Exporters = exporters
	device = viper.GetString("device")
	return nil
}
