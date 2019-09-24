package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type pubsubExporter struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// New creates a new Google Pub/Sub reporter
func New() (exporter.Exporter, error) {
	projectName := os.Getenv("RUUVITAG_PUBSUB_PROJECT")
	if projectName == "" {
		return nil, fmt.Errorf("RUUVITAG_PUBSUB_PROJECT must be set")
	}
	client, err := pubsub.NewClient(context.Background(), projectName)
	if err != nil {
		return nil, err
	}
	topicName := os.Getenv("RUUVITAG_PUBSUB_TOPIC")
	if topicName == "" {
		return nil, fmt.Errorf("RUUVITAG_PUBSUB_TOPIC must be set")
	}
	topic, err := client.CreateTopic(context.Background(), topicName)
	if err != nil {
		return nil, err
	}
	return &pubsubExporter{
		client: client,
		topic:  topic,
	}, nil
}

func (e *pubsubExporter) Name() string {
	return "Google Pub/Sub"
}

func (e *pubsubExporter) Export(ctx context.Context, data ruuvitag.SensorData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := &pubsub.Message{
		Data: jsonData,
		Attributes: map[string]string{
			"mac":  data.DeviceID,
			"name": data.Name,
		},
	}
	e.topic.Publish(ctx, msg)
	return nil
}

func (e *pubsubExporter) Close() error {
	e.topic.Stop()
	return e.client.Close()
}
