package sqs

import (
	"fmt"
	"log/slog"
)

type Config struct {
	QueueName       string
	QueueURL        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Columns         map[string]string
	Logger          *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.QueueName == "" && cfg.QueueURL == "" {
		return fmt.Errorf("AWS SQS queue name or queue URL must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}
