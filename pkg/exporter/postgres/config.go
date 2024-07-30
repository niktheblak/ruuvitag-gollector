package postgres

import (
	"log/slog"
)

type Config struct {
	ConnString string
	Table      string
	Columns    map[string]string
	Logger     *slog.Logger
}
