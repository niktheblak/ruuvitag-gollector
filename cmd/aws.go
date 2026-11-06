//go:build aws

package cmd

import (
	"context"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/dynamodb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/sqs"
)

func createDynamoDBExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exporterInitTimeout)
	defer cancel()
	return dynamodb.New(ctx, name, dynamodb.Config{
		Table:           cast.ToString(cfg["table"]),
		Region:          cast.ToString(cfg["region"]),
		AccessKeyID:     cast.ToString(cfg["access_key_id"]),
		SecretAccessKey: cast.ToString(cfg["secret_access_key"]),
		SessionToken:    cast.ToString(cfg["session_token"]),
		Columns:         columns,
		Logger:          logger,
	})
}

func createSQSExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exporterInitTimeout)
	defer cancel()
	return sqs.New(ctx, name, sqs.Config{
		QueueName:       cast.ToString(cfg["queue_name"]),
		QueueURL:        cast.ToString(cfg["queue_url"]),
		Region:          cast.ToString(cfg["region"]),
		AccessKeyID:     cast.ToString(cfg["access_key_id"]),
		SecretAccessKey: cast.ToString(cfg["secret_access_key"]),
		SessionToken:    cast.ToString(cfg["session_token"]),
		Columns:         columns,
		Logger:          logger,
	})
}
