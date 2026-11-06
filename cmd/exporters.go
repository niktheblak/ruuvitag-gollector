package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"

	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/console"
	"github.com/niktheblak/ruuvitag-gollector/pkg/scanner"
)

func createExporters() error {
	if viper.ConfigFileUsed() != "" {
		logger.LogAttrs(context.TODO(), slog.LevelInfo, "Read config from file", slog.String("file", viper.ConfigFileUsed()))
	}
	ruuviTags := viper.GetStringMapString("ruuvitags")
	if len(ruuviTags) == 0 {
		return fmt.Errorf("at least one RuuviTag address must be specified")
	}
	logger.LogAttrs(context.TODO(), slog.LevelInfo, "RuuviTags", slog.Any("ruuvitags", ruuviTags))
	columns := viper.GetStringMapString("columns")
	if len(columns) == 0 {
		columns = sensor.DefaultColumnMap
	}
	if err := sensor.ValidateColumnMapping(columns); err != nil {
		return fmt.Errorf("invalid column mapping: %w", err)
	}
	logger.LogAttrs(context.TODO(), slog.LevelInfo, "Using column mapping", slog.Any("columns", columns))
	peripherals = make(map[string]string)
	for addr, name := range ruuviTags {
		peripherals[scanner.NormalizeAddress(addr)] = name
	}
	exporterConfigs, err := getExporterConfigs()
	if err != nil {
		return err
	}
	for name, cfg := range exporterConfigs {
		exp, err := createExporter(name, columns, cfg)
		if err != nil {
			return errors.Join(err, closeExporters())
		}
		exporters = append(exporters, exp)
	}
	device = viper.GetString(deviceConfigKey)
	logger.Info("Using device", "device", device)
	return nil
}

func createExporter(name string, columns map[string]string, cfg map[string]any) (exp exporter.Exporter, err error) {
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
		exp, err = createInfluxDBExporter(name, columns, cfg)
	case "console":
		exp = console.New(name)
	case "pubsub":
		exp, err = createPubSubExporter(name, columns, cfg)
	case "dynamodb":
		exp, err = createDynamoDBExporter(name, columns, cfg)
	case "sqs":
		exp, err = createSQSExporter(name, columns, cfg)
	case "postgres":
		exp, err = createPostgresExporter(name, columns, cfg)
	case "http":
		exp, err = createHTTPExporter(name, columns, cfg)
	case "mqtt":
		exp, err = createMQTTExporter(name, columns, cfg)
	default:
		err = fmt.Errorf("invalid exporter type: %s", tp)
	}
	if err != nil {
		err = fmt.Errorf("failed to create exporter %s: %w", name, err)
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
	exporters = nil
	return errors.Join(errs...)
}

func getExporterConfigs() (map[string]map[string]any, error) {
	cfg, err := parseExporterConfig()
	if err != nil {
		return nil, err
	}
	oldCfg := parseOldExporterConfig()
	if len(oldCfg) > 0 {
		logger.Warn("Using deprecated exporter config. Consult README.md for the new configuration format.")
	}
	for k, v := range oldCfg {
		if _, exists := cfg[k]; exists {
			return nil, fmt.Errorf("exporter %s is configured both in the exporters table and in the deprecated top-level config", k)
		}
		cfg[k] = v
	}
	return cfg, nil
}

func parseExporterConfig() (map[string]map[string]any, error) {
	configs := make(map[string]map[string]any)
	raw := viper.Get("exporters")
	if raw == nil {
		return configs, nil
	}
	cfgMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid exporters config: %v", raw)
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
		configs["postgres"] = withType(postgresCfg, "postgres")
	}
	influxDBCfg, ok := viper.Get("influxdb").(map[string]any)
	if ok && isEnabled(influxDBCfg) {
		configs["influxdb"] = withType(influxDBCfg, "influxdb")
	}
	dynamoDBCfg, ok := viper.Get("aws.dynamodb").(map[string]any)
	if ok && isEnabled(dynamoDBCfg) {
		dynamoDBCfg = withType(dynamoDBCfg, "dynamodb")
		dynamoDBCfg["access_key_id"] = viper.GetString("aws.access_key_id")
		dynamoDBCfg["secret_access_key"] = viper.GetString("aws.secret_access_key")
		dynamoDBCfg["session_token"] = viper.GetString("aws.session_token")
		dynamoDBCfg["region"] = viper.GetString("aws.region")
		configs["dynamodb"] = dynamoDBCfg
	}
	sqsCfg, ok := viper.Get("aws.sqs").(map[string]any)
	if ok && isEnabled(sqsCfg) {
		sqsCfg = withType(sqsCfg, "sqs")
		delete(sqsCfg, "queue")
		if queueName := viper.GetString("aws.sqs.queue.name"); queueName != "" {
			sqsCfg["queue_name"] = queueName
		}
		if queueURL := viper.GetString("aws.sqs.queue.url"); queueURL != "" {
			sqsCfg["queue_url"] = queueURL
		}
		sqsCfg["access_key_id"] = viper.GetString("aws.access_key_id")
		sqsCfg["secret_access_key"] = viper.GetString("aws.secret_access_key")
		sqsCfg["session_token"] = viper.GetString("aws.session_token")
		sqsCfg["region"] = viper.GetString("aws.region")
		configs["sqs"] = sqsCfg
	}
	gcpCfg, ok := viper.Get("gcp").(map[string]any)
	if ok && isEnabled(gcpCfg) {
		gcpCfg = withType(gcpCfg, "pubsub")
		gcpCfg["credentials"] = viper.GetString("gcp.credentials")
		gcpCfg["project"] = viper.GetString("gcp.project")
		if topic := viper.GetString("gcp.pubsub.topic"); topic != "" {
			gcpCfg["topic"] = topic
		}
		configs["pubsub"] = gcpCfg
	}
	mqttCfg, ok := viper.Get("mqtt").(map[string]any)
	if ok && isEnabled(mqttCfg) {
		configs["mqtt"] = withType(mqttCfg, "mqtt")
	}
	httpCfg, ok := viper.Get("http").(map[string]any)
	if ok && isEnabled(httpCfg) {
		configs["http"] = withType(httpCfg, "http")
	}
	if viper.GetBool("console") {
		configs["console"] = map[string]any{"type": "console"}
	}
	return configs
}

// withType returns a copy of the given deprecated exporter config with the exporter type filled in
// unless it is already set.
func withType(cfg map[string]any, tp string) map[string]any {
	cfg = maps.Clone(cfg) // defensive copy
	if _, ok := cfg["type"]; !ok {
		cfg["type"] = tp
	}
	return cfg
}

func isEnabled(cfg map[string]any) bool {
	enabled, ok := cfg["enabled"].(bool)
	if !ok {
		return true
	}
	return enabled
}
