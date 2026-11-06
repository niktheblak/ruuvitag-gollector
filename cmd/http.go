package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/http"
)

func createHTTPExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx := context.Background()
	url := cast.ToString(cfg["url"])
	if url == "" {
		url = cast.ToString(cfg["addr"])
	}
	timeout, err := configDuration(cfg["timeout"])
	if err != nil {
		return nil, err
	}
	if timeout <= 0 {
		timeout = http.DefaultTimeout
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "Using HTTP endpoint", slog.String("url", url), slog.Duration("timeout", timeout))
	return http.New(ctx, name, http.Config{
		URL:     url,
		Token:   cast.ToString(cfg["token"]),
		Timeout: timeout,
		Columns: columns,
		Logger:  logger,
	})
}
