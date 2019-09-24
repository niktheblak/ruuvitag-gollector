package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/niktheblak/ruuvitag-gollector/pkg/reporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type pubsubReporter struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// New creates a new Google Pub/Sub reporter
func New() (reporter.Reporter, error) {
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
	return &pubsubReporter{
		client: client,
		topic:  topic,
	}, nil
}

func (r *pubsubReporter) Name() string {
	return "Google Pub/Sub"
}

func (r *pubsubReporter) Report(data ruuvitag.SensorData) error {
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
	r.topic.Publish(context.Background(), msg)
	return nil
}

func (r *pubsubReporter) Close() error {
	r.topic.Stop()
	return r.client.Close()
}
