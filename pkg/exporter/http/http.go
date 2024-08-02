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

type Config struct {
	URL     string
	Token   string
	Timeout time.Duration
	Columns map[string]string
	Logger  *slog.Logger
}

type httpExporter struct {
	client  *nethttp.Client
	url     string
	token   string
	columns map[string]string
	logger  *slog.Logger
}

func New(cfg Config) (exporter.Exporter, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("parameter url must be non-empty")
	}
	if len(cfg.Columns) == 0 {
		return nil, fmt.Errorf("columns must be non-empty")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "HTTP")
	client := &nethttp.Client{
		Timeout: cfg.Timeout,
	}
	return &httpExporter{
		client:  client,
		url:     cfg.URL,
		token:   cfg.Token,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (h *httpExporter) Name() string {
	return fmt.Sprintf("HTTP (%s)", h.url)
}

func (h *httpExporter) Export(ctx context.Context, data sensor.Data) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(columnmap.Transform(h.columns, data))
	if err != nil {
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
	content, _ := io.ReadAll(resp.Body)
	h.logger.LogAttrs(ctx, slog.LevelDebug, "Server response", slog.Int("status", resp.StatusCode), slog.String("body", string(content)))
	return resp.Body.Close()
}

func (h *httpExporter) Close() error {
	h.client.CloseIdleConnections()
	return nil
}
