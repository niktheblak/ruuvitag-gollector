//go:build mqtt

package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

const (
	topicPrefix       = "ruuvitag-gollector"
	disconnectQuiesce = 250 * time.Millisecond
)

type mqttExporter struct {
	name    string
	client  mqtt.Client
	columns map[string]string
	logger  *slog.Logger
}

func New(ctx context.Context, name string, cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	cfg.Logger = cfg.Logger.With("exporter", "MQTT")
	if cfg.ClientId == "" {
		cfg.ClientId = DefaultClientID
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Addr)
	opts.SetClientID(cfg.ClientId)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	opts.SetAutoReconnect(cfg.AutoReconnect)
	if cfg.ReconnectInterval > 0 {
		opts.SetMaxReconnectInterval(cfg.ReconnectInterval)
	}
	tlsConfig, err := newTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}
	logger := cfg.Logger
	opts.SetOnConnectHandler(func(_ mqtt.Client) {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "Connected to MQTT broker", slog.String("addr", cfg.Addr))
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "Lost connection to MQTT broker", slog.String("addr", cfg.Addr), slog.Any("error", err))
	})
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to MQTT broker", slog.String("addr", cfg.Addr), slog.String("client_id", cfg.ClientId))
	client := mqtt.NewClient(opts)
	if err := wait(ctx, client.Connect()); err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker %s: %w", cfg.Addr, err)
	}
	return &mqttExporter{
		name:    name,
		client:  client,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func newTLSConfig(cfg Config) (*tls.Config, error) {
	if cfg.CaFile == "" {
		return nil, nil
	}
	ca, err := os.ReadFile(cfg.CaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file %s: %w", cfg.CaFile, err)
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("no valid certificates found in CA file %s", cfg.CaFile)
	}
	return &tls.Config{
		RootCAs: certpool,
	}, nil
}

// wait blocks until the token completes or the context is done.
func wait(ctx context.Context, token mqtt.Token) error {
	select {
	case <-token.Done():
		return token.Error()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Topic returns the MQTT topic a measurement from the given tag is published to.
func Topic(data sensor.Data) string {
	mac := strings.ReplaceAll(data.Addr, ":", "")
	return fmt.Sprintf("%s/%s/%s", topicPrefix, data.Name, mac)
}

func (e *mqttExporter) Name() string {
	return e.name
}

func (e *mqttExporter) Export(ctx context.Context, data sensor.Data) error {
	payload, err := json.Marshal(columnmap.Transform(e.columns, data))
	if err != nil {
		return err
	}
	topic := Topic(data)
	e.logger.LogAttrs(ctx, slog.LevelDebug, "Publishing measurement", slog.String("topic", topic), slog.String("payload", string(payload)))
	return wait(ctx, e.client.Publish(topic, 0, false, payload))
}

func (e *mqttExporter) Close() error {
	e.client.Disconnect(uint(disconnectQuiesce / time.Millisecond))
	return nil
}
