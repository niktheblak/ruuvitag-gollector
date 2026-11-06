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

func createPubSubExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exporterInitTimeout)
	defer cancel()
	var credentialsJSON []byte
	credentialsFile := cast.ToString(cfg["credentials"])
	if credentialsFile != "" {
		var err error
		credentialsJSON, err = os.ReadFile(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read Google credentials file %s: %w", credentialsFile, err)
		}
	}
	return pubsub.New(ctx, name, pubsub.Config{
		Project:         cast.ToString(cfg["project"]),
		Topic:           cast.ToString(cfg["topic"]),
		CredentialsJSON: credentialsJSON,
		Columns:         columns,
		Logger:          logger,
	})
}
