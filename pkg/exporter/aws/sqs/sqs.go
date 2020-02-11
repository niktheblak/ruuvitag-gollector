package sqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awssqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Config struct {
	QueueName       string
	QueueURL        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type sqsExporter struct {
	sess     *session.Session
	sqs      sqsiface.SQSAPI
	queueUrl string
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
	sqs := awssqs.New(sess)
	var queueUrl string
	if cfg.QueueURL != "" {
		queueUrl = cfg.QueueURL
	} else {
		resp, err := sqs.GetQueueUrl(&awssqs.GetQueueUrlInput{
			QueueName: aws.String(cfg.QueueName),
		})
		if err != nil {
			return nil, err
		}
		queueUrl = *resp.QueueUrl
	}
	return &sqsExporter{
		sess:     sess,
		sqs:      sqs,
		queueUrl: queueUrl,
	}, nil
}

func (e *sqsExporter) Name() string {
	return "AWS SQS"
}

func (e *sqsExporter) Export(ctx context.Context, data sensor.Data) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	input := &awssqs.SendMessageInput{
		MessageAttributes: map[string]*awssqs.MessageAttributeValue{
			"mac": {
				DataType:    aws.String("String"),
				StringValue: aws.String(data.Addr),
			},
			"name": {
				DataType:    aws.String("String"),
				StringValue: aws.String(data.Name),
			},
		},
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(e.queueUrl),
	}
	_, err = e.sqs.SendMessageWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (e *sqsExporter) Close() error {
	return nil
}
