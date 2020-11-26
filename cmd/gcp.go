// +build gcp

package cmd

import (
	"context"
	"fmt"

	"github.com/niktheblak/gcloudzap"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/gcp/pubsub"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().Bool("gcp.stackdriver.enabled", false, "Send logs to Google Stackdriver")
	rootCmd.PersistentFlags().String("gcp.credentials", "", "Google Cloud application credentials file")
	rootCmd.MarkFlagFilename("gcp.credentials", "json")
	rootCmd.PersistentFlags().String("gcp.project", "", "Google Cloud Platform project")
	rootCmd.PersistentFlags().Bool("gcp.pubsub.enabled", false, "Send measurements to Google Pub/Sub")
	rootCmd.PersistentFlags().String("gcp.pubsub.topic", "", "Google Pub/Sub topic to use")
}

func initStackdriverLogging() error {
	project := viper.GetString("gcp.project")
	if project == "" {
		return fmt.Errorf("Google Cloud Platform project must be specified")
	}
	var err error
	logger, err = gcloudzap.NewProduction(project, "ruuvitag-gollector")
	return err
}

func addPubSubExporter(exporters *[]exporter.Exporter) error {
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
	*exporters = append(*exporters, ps)
	return nil
}
