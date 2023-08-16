//go:build gcp

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/gcp/pubsub"
)

func init() {
	rootCmd.PersistentFlags().String("gcp.credentials", "", "Google Cloud application credentials file")
	if err := rootCmd.MarkFlagFilename("gcp.credentials", "json"); err != nil {
		panic(err)
	}
	rootCmd.PersistentFlags().String("gcp.project", "", "Google Cloud Platform project")
	rootCmd.PersistentFlags().Bool("gcp.pubsub.enabled", false, "Send measurements to Google Pub/Sub")
	rootCmd.PersistentFlags().String("gcp.pubsub.topic", "", "Google Pub/Sub topic to use")
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
		return err
	}
	*exporters = append(*exporters, ps)
	return nil
}
