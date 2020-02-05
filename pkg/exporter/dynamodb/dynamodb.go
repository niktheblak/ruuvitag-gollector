package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type AWSData struct {
	Addr            string    `dynamodbav:"Addr"`
	Name            string    `dynamodbav:"Name"`
	Temperature     float64   `dynamodbav:"Temperature"`
	Humidity        float64   `dynamodbav:"Humidity"`
	Pressure        float64   `dynamodbav:"Pressure"`
	Battery         int       `dynamodbav:"Battery"`
	AccelerationX   int       `dynamodbav:"AccelerationX"`
	AccelerationY   int       `dynamodbav:"AccelerationY"`
	AccelerationZ   int       `dynamodbav:"AccelerationZ"`
	MovementCounter int       `dynamodbav:"MovementCounter"`
	Timestamp       time.Time `dynamodbav:"Timestamp"`
}

func From(data sensor.Data) AWSData {
	return AWSData{
		Addr:            data.Addr,
		Name:            data.Name,
		Temperature:     data.Temperature,
		Humidity:        data.Humidity,
		Pressure:        data.Pressure,
		Battery:         data.Battery,
		AccelerationX:   data.AccelerationX,
		AccelerationY:   data.AccelerationY,
		AccelerationZ:   data.AccelerationZ,
		MovementCounter: data.MovementCounter,
		Timestamp:       data.Timestamp,
	}
}

type Config struct {
	Table           string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type dynamoDBExporter struct {
	sess  *session.Session
	db    *dynamodb.DynamoDB
	table string
}

func New(cfg Config) (exporter.Exporter, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		MaxRetries:  aws.Int(3),
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
	item, err := dynamodbattribute.MarshalMap(From(data))
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
