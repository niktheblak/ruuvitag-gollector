//go:build aws

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type dynamoDBExporter struct {
	sess  *session.Session
	db    dynamodbiface.DynamoDBAPI
	table string
}

func New(cfg Config) (exporter.Exporter, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
	})
	if err != nil {
		return nil, err
	}
	db := dynamodb.New(sess)
	return &dynamoDBExporter{
		sess:  sess,
		db:    db,
		table: cfg.Table,
	}, nil
}

func (e *dynamoDBExporter) Name() string {
	return "AWS DynamoDB"
}

func (e *dynamoDBExporter) Export(ctx context.Context, data sensor.Data) error {
	item, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(e.table),
	}
	_, err = e.db.PutItemWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (e *dynamoDBExporter) Close() error {
	return nil
}
