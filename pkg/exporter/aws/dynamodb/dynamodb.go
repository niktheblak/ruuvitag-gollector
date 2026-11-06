//go:build aws

package dynamodb

import (
	"context"
	"log/slog"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws"
)

type dynamoDBExporter struct {
	name    string
	db      dynamodbiface.DynamoDBAPI
	table   string
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
	cfg.Logger = cfg.Logger.With("exporter", "AWS DynamoDB")
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to DynamoDB", slog.String("region", cfg.Region), slog.String("table", cfg.Table))
	sess, err := aws.NewSession(aws.Credentials{
		Region:          cfg.Region,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		SessionToken:    cfg.SessionToken,
	}, cfg.Logger)
	if err != nil {
		return nil, err
	}
	return &dynamoDBExporter{
		name:    name,
		db:      dynamodb.New(sess),
		table:   cfg.Table,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (e *dynamoDBExporter) Name() string {
	return e.name
}

func (e *dynamoDBExporter) Export(ctx context.Context, data sensor.Data) error {
	item, err := dynamodbattribute.MarshalMap(columnmap.Transform(e.columns, data))
	if err != nil {
		return err
	}
	e.logger.LogAttrs(ctx, slog.LevelDebug, "Putting item", slog.String("table", e.table), slog.Any("item", item))
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: awssdk.String(e.table),
	}
	_, err = e.db.PutItemWithContext(ctx, input)
	return err
}

func (e *dynamoDBExporter) Close() error {
	return nil
}
