package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	nethttp "net/http"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

const (
	DefaultTimeout = 10 * time.Second

	// maxErrorBodyLength limits how much of an error response body is included in the returned error.
	maxErrorBodyLength = 1024
)

type Config struct {
	URL     string
	Token   string
	Timeout time.Duration
	Columns map[string]string
	Logger  *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.URL == "" {
		return fmt.Errorf("URL must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}

type httpExporter struct {
	name    string
	client  *nethttp.Client
	url     string
	token   string
	columns map[string]string
	logger  *slog.Logger
}

func New(_ context.Context, name string, cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	cfg.Logger = cfg.Logger.With("exporter", "HTTP")
	client := &nethttp.Client{
		Timeout: cfg.Timeout,
	}
	return &httpExporter{
		name:    name,
		client:  client,
		url:     cfg.URL,
		token:   cfg.Token,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (h *httpExporter) Name() string {
	return h.name
}

func (h *httpExporter) Export(ctx context.Context, data sensor.Data) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(columnmap.Transform(h.columns, data)); err != nil {
		return err
	}
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, h.url, buf)
	if err != nil {
		return err
	}
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Sending measurement", slog.String("url", h.url), slog.String("data", buf.String()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("From", "ruuvitag-gollector")
	if h.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.token))
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyLength))
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Server response", slog.Int("status", resp.StatusCode), slog.String("body", string(content)))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("server returned status %s: %s", resp.Status, bytes.TrimSpace(content))
	}
	return nil
}

func (h *httpExporter) Close() error {
	h.client.CloseIdleConnections()
	return nil
}
