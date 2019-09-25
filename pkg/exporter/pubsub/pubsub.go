package pubsub

import (
	"context"
	"encoding/json"

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
	client, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	t, err := client.CreateTopic(ctx, topic)
	if err != nil {
		return nil, err
	}
	return &pubsubExporter{
		client: client,
		topic:  t,
	}, nil
}

func (e *pubsubExporter) Name() string {
	return "Google Pub/Sub"
}

func (e *pubsubExporter) Export(ctx context.Context, data sensor.Data) error {
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
