package dynamodb

import (
	"fmt"
	"log/slog"
)

type Config struct {
	Table           string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Columns         map[string]string
	Logger          *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.Table == "" {
		return fmt.Errorf("DynamoDB table name must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}
