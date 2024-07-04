//go:build gcp

package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type pubsubExporter struct {
	client  *pubsub.Client
	topic   *pubsub.Topic
	columns map[string]string
	logger  *slog.Logger
}

// New creates a new Google Pub/Sub reporter
func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
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
	t := client.Topic(cfg.Topic)
	return &pubsubExporter{
		client:  client,
		topic:   t,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (e *pubsubExporter) Name() string {
	return "Google Pub/Sub"
}

func (e *pubsubExporter) Export(ctx context.Context, data sensor.Data) error {
	fields := make(map[string]any)
	columnmap.Collect(e.columns, data, func(column string, v any) {
		switch column {
		case "mac":
			break
		case "name":
			break
		default:
			fields[column] = v
		}
	})
	jsonData, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	e.logger.LogAttrs(ctx, slog.LevelInfo, "Publishing measurement", slog.String("data", string(jsonData)), slog.String("mac", data.Addr), slog.String("name", data.Name))
	msg := &pubsub.Message{
		Data: jsonData,
		Attributes: map[string]string{
			"mac":  data.Addr,
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
