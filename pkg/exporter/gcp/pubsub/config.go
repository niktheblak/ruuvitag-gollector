package pubsub

import (
	"fmt"
	"log/slog"
)

type Config struct {
	Project         string
	Topic           string
	CredentialsJSON []byte
	Columns         map[string]string
	Logger          *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.Project == "" {
		return fmt.Errorf("GCP project must be specified")
	}
	if cfg.Topic == "" {
		return fmt.Errorf("Pub/Sub topic must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}
