//go:build gcp

package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type pubsubExporter struct {
	name    string
	client  *pubsub.Client
	topic   *pubsub.Topic
	columns map[string]string
	logger  *slog.Logger
}

// New creates a new Google Pub/Sub exporter
func New(ctx context.Context, name string, cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	cfg.Logger = cfg.Logger.With("exporter", "Google Pub/Sub")
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to Pub/Sub", slog.String("project", cfg.Project), slog.String("topic", cfg.Topic))
	var opts []option.ClientOption
	if len(cfg.CredentialsJSON) > 0 {
		cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Using credentials JSON")
		opts = append(opts, option.WithCredentialsJSON(cfg.CredentialsJSON))
	}
	client, err := pubsub.NewClient(ctx, cfg.Project, opts...)
	if err != nil {
		return nil, fmt.Errorf("error while creating client: %w", err)
	}
	return &pubsubExporter{
		name:    name,
		client:  client,
		topic:   client.Topic(cfg.Topic),
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (e *pubsubExporter) Name() string {
	return e.name
}

func (e *pubsubExporter) Export(ctx context.Context, data sensor.Data) error {
	fields := columnmap.Transform(e.columns, data)
	attrs := make(map[string]string)
	// mac and name are sent as message attributes, so they are removed from the body
	if mac, ok := e.columns["mac"]; ok {
		delete(fields, mac)
		attrs[mac] = data.Addr
	}
	if name, ok := e.columns["name"]; ok {
		delete(fields, name)
		attrs[name] = data.Name
	}
	jsonData, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	e.logger.LogAttrs(ctx, slog.LevelDebug, "Publishing measurement", slog.String("data", string(jsonData)), slog.Any("attributes", attrs))
	msg := &pubsub.Message{
		Data:       jsonData,
		Attributes: attrs,
	}
	_, err = e.topic.Publish(ctx, msg).Get(ctx)
	return err
}

func (e *pubsubExporter) Close() error {
	e.topic.Stop()
	return e.client.Close()
}
