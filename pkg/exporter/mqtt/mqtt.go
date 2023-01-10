//go:build mqtt

package mqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type mqttExporter struct {
	client    mqtt.Client
	tlsConfig *tls.Config
}

func New(cfg Config) (exporter.Exporter, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Addr)
	opts.SetClientID(cfg.ClientId)
	if cfg.Username != "" && cfg.Password != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}
	opts.SetAutoReconnect(cfg.AutoReconnect)
	opts.SetMaxReconnectInterval(cfg.ReconnectInterval)
	tlsConfig, err := newTlsConfig(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &mqttExporter{
		client:    client,
		tlsConfig: tlsConfig,
	}, nil
}

func newTlsConfig(cfg Config) (*tls.Config, error) {
	if cfg.CaFile != "" {
		certpool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cfg.CaFile)
		if err != nil {
			return nil, err
		}
		certpool.AppendCertsFromPEM(ca)
		return &tls.Config{
			RootCAs: certpool,
		}, nil
	}
	return nil, nil
}

func (m mqttExporter) Name() string {
	return fmt.Sprintf("MQTT")
}

func (m mqttExporter) Export(ctx context.Context, data sensor.Data) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return err
	}
	mac := strings.Replace(data.Addr, ":", "", -1)
	topic := fmt.Sprintf("ruuvitag-gollector/%s/%s", data.Name, mac)
	token := m.client.Publish(topic, 0, false, buf.String())
	token.Wait()
	return token.Error()
}

func (m mqttExporter) Close() error {
	m.client.Disconnect(0)
	return nil
}
