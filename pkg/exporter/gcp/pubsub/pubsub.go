// +build gcp

package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type pubsubExporter struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// New creates a new Google Pub/Sub reporter
func New(ctx context.Context, project, topic string) (exporter.Exporter, error) {
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if creds == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS must be set")
	}
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("error while creating client: %w", err)
	}
	t := client.Topic(topic)
	return &pubsubExporter{
		client: client,
		topic:  t,
	}, nil
}

func (e *pubsubExporter) Name() string {
	return "Google Pub/Sub"
}

func (e *pubsubExporter) Export(ctx context.Context, data sensor.Data) error {
	data.Addr = strings.ToUpper(data.Addr)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := &pubsub.Message{
		Data: jsonData,
		Attributes: map[string]string{
			"mac":  strings.ToUpper(data.Addr),
			"name": data.Name,
		},
	}
	_, err = e.topic.Publish(ctx, msg).Get(ctx)
	return err
}

func (e *pubsubExporter) Close() error {
	e.topic.Stop()
	return e.client.Close()
}
