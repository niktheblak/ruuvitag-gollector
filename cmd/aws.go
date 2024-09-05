//go:build aws

package cmd

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/dynamodb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/sqs"
)

func createDynamoDBExporter(cfg map[string]any) (exporter.Exporter, error) {
	table := cast.ToString(cfg["dynamodb.table"])
	if table == "" {
		return nil, fmt.Errorf("DynamoDB table name must be specified")
	}
	return dynamodb.New(dynamodb.Config{
		Table:           table,
		Region:          cast.ToString(cfg["region"]),
		AccessKeyID:     cast.ToString(cfg["access_key_id"]),
		SecretAccessKey: cast.ToString(cfg["secret_access_key"]),
		SessionToken:    cast.ToString(cfg["session_token"]),
	})
}

func createSQSExporter(cfg map[string]any) (exporter.Exporter, error) {
	queueName := cast.ToString(cfg["sqs.queue.name"])
	queueURL := cast.ToString(cfg["sqs.queue.url"])
	if queueName == "" && queueURL == "" {
		return nil, fmt.Errorf("AWS SQS queue name or queue URL must be specified")
	}
	return sqs.New(sqs.Config{
		QueueName:       queueName,
		QueueURL:        queueURL,
		Region:          cast.ToString(cfg["region"]),
		AccessKeyID:     cast.ToString(cfg["access_key_id"]),
		SecretAccessKey: cast.ToString(cfg["secret_access_key"]),
		SessionToken:    cast.ToString(cfg["session_token"]),
	})
}
