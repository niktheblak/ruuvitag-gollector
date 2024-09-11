package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/http"
)

func createExporters() error {
	if viper.ConfigFileUsed() != "" {
		logger.LogAttrs(nil, slog.LevelInfo, "Read config from file", slog.String("file", viper.ConfigFileUsed()))
	}
	ruuviTags := viper.GetStringMapString("ruuvitags")
	if len(ruuviTags) == 0 {
		return fmt.Errorf("at least one RuuviTag address must be specified")
	}
	logger.LogAttrs(nil, slog.LevelInfo, "RuuviTags", slog.Any("ruuvitags", ruuviTags))
	columns := viper.GetStringMapString("columns")
	if columns == nil || len(columns) == 0 {
		columns = sensor.DefaultColumnMap
	}
	logger.LogAttrs(nil, slog.LevelInfo, "Using column mapping", slog.Any("columns", columns))
	peripherals = make(map[string]string)
	for addr, name := range ruuviTags {
		peripherals[ble.NewAddr(addr).String()] = name
	}
	exporterConfigs, err := getExporterConfigs()
	if err != nil {
		return err
	}
	for name, cfg := range exporterConfigs {
		exp, err := createExporter(name, cfg, columns)
		if err != nil {
			return err
		}
		exporters = append(exporters, exp)
	}
	device = viper.GetString("device")
	logger.Info("Using device", "device", device)
	return nil
}

func createExporter(name string, cfg map[string]any, columns map[string]string) (exp exporter.Exporter, err error) {
	logger := logger.With("name", name)
	rawType, ok := cfg["type"]
	if !ok {
		err = fmt.Errorf("exporter type is not specified in config: %v", cfg)
		return
	}
	tp, err := cast.ToStringE(rawType)
	if err != nil {
		err = fmt.Errorf("exporter type is not a string: %w", err)
		return
	}
	logger = logger.With("type", tp)
	logger.Info("Creating exporter")
	switch tp {
	case "influxdb":
		exp, err = createInfluxDBExporter(columns, cfg)
	case "console":
		exp = console.New(name)
	case "pubsub":
		exp, err = createPubSubExporter(columns, cfg)
	case "dynamodb":
		exp, err = createDynamoDBExporter(cfg)
	case "sqs":
		exp, err = createSQSExporter(cfg)
	case "postgres":
		exp, err = createPostgresExporter(name, columns, cfg)
	case "http":
		addr := cast.ToString(cfg["addr"])
		token := cast.ToString(cfg["token"])
		exp, err = http.New(http.Config{
			URL:     addr,
			Token:   token,
			Timeout: 10 * time.Second,
			Columns: columns,
			Logger:  logger,
		})
	case "mqtt":
		exp, err = createMQTTExporter(cfg)
	default:
		err = fmt.Errorf("invalid exporter type: %s", tp)
	}
	if err != nil {
		err = fmt.Errorf("failed to create exporter: %w", err)
	}
	return
}

func closeExporters() error {
	var errs []error
	for _, exp := range exporters {
		if err := exp.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func getExporterConfigs() (map[string]map[string]any, error) {
	cfg, err := parseExporterConfig()
	if err != nil {
		return nil, err
	}
	oldCfg := parseOldExporterConfig()
	for k, v := range oldCfg {
		cfg[k] = v
	}
	return cfg, nil
}

func parseExporterConfig() (map[string]map[string]any, error) {
	configs := make(map[string]map[string]any)
	cfgMap, ok := viper.Get("exporters").(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid exporters config: %v", viper.Get("exporters"))
	}
	for name, cfg := range cfgMap {
		values, ok := cfg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid config for exporter %s: %v", name, cfg)
		}
		configs[name] = values
	}
	return configs, nil
}

func parseOldExporterConfig() map[string]map[string]any {
	configs := make(map[string]map[string]any)
	postgresCfg, ok := viper.Get("postgres").(map[string]any)
	if ok && isEnabled(postgresCfg) {
		configs["postgres"] = postgresCfg
	}
	influxDBCfg, ok := viper.Get("influxdb").(map[string]any)
	if ok && isEnabled(influxDBCfg) {
		configs["influxdb"] = influxDBCfg
	}
	dynamoDBCfg, ok := viper.Get("aws.dynamodb").(map[string]any)
	if ok && isEnabled(dynamoDBCfg) {
		configs["dynamodb"] = dynamoDBCfg
	}
	sqsCfg, ok := viper.Get("aws.sqs").(map[string]any)
	if ok && isEnabled(sqsCfg) {
		configs["sqs"] = sqsCfg
	}
	gcpCfg, ok := viper.Get("gcp").(map[string]any)
	if ok && isEnabled(gcpCfg) {
		configs["gcp"] = gcpCfg
	}
	mqttCfg, ok := viper.Get("mqtt").(map[string]any)
	if ok && isEnabled(mqttCfg) {
		configs["mqtt"] = mqttCfg
	}
	return configs
}

func isEnabled(cfg map[string]any) bool {
	enabled, ok := cfg["enabled"].(bool)
	if !ok {
		return true
	}
	return enabled
}
