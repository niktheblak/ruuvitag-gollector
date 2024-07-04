package pubsub

import (
	"log/slog"
)

type Config struct {
	Project         string
	Topic           string
	CredentialsJSON []byte
	Columns         map[string]string
	Logger          *slog.Logger
}
