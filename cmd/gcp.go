//go:build gcp

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/gcp/pubsub"
)

func createPubSubExporter(columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx := context.Background()
	project := cast.ToString(cfg["project"])
	if project == "" {
		return nil, fmt.Errorf("Google Cloud Platform project must be specified")
	}
	topic := cast.ToString(cfg["topic"])
	if topic == "" {
		return nil, fmt.Errorf("Google Pub/Sub topic must be specified")
	}
	var credentialsJSON []byte
	credentialsFile := cast.ToString(cfg["credentials"])
	var err error
	if credentialsFile != "" {
		credentialsJSON, err = os.ReadFile(credentialsFile)
		if err != nil {
			return nil, err
		}
	}
	return pubsub.New(ctx, pubsub.Config{
		Project:         project,
		Topic:           topic,
		CredentialsJSON: credentialsJSON,
		Columns:         columns,
		Logger:          logger,
	})
}
