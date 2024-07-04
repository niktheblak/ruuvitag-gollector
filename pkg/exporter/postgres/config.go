package postgres

import (
	"log/slog"
)

type Config struct {
	PSQLInfo string
	Table    string
	Columns  map[string]string
	Logger   *slog.Logger
}
