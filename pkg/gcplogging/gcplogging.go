package gcplogging

import (
	"context"
	"fmt"

	gcplogging "cloud.google.com/go/logging"
	"github.com/op/go-logging"
)

type GCPBackend struct {
	Client *gcplogging.Client
	Logger *gcplogging.Logger
}

func (b *GCPBackend) Log(level logging.Level, calldepth int, rec *logging.Record) error {
	b.Logger.Log(toEntry(rec))
	return nil
}

func (b *GCPBackend) Close() error {
	return b.Client.Close()
}

func NewBackend(project, logID string, opts ...gcplogging.LoggerOption) (*GCPBackend, error) {
	ctx := context.Background()
	client, err := gcplogging.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stackdriver client: %w", err)
	}
	return &GCPBackend{
		Client: client,
		Logger: client.Logger(logID, opts...),
	}, nil
}

func toEntry(r *logging.Record) gcplogging.Entry {
	return gcplogging.Entry{
		Timestamp: r.Time,
		Severity:  toSeverity(r.Level),
		Payload:   r.Message(),
	}
}

func toSeverity(level logging.Level) gcplogging.Severity {
	switch level {
	case logging.CRITICAL:
		return gcplogging.Critical
	case logging.ERROR:
		return gcplogging.Error
	case logging.WARNING:
		return gcplogging.Warning
	case logging.NOTICE:
		return gcplogging.Notice
	case logging.INFO:
		return gcplogging.Info
	case logging.DEBUG:
		return gcplogging.Debug
	default:
		return gcplogging.Default
	}
}
