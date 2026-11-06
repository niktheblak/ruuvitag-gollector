//go:build aws

package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws"
)

const fifoSuffix = ".fifo"

// sendMessageAPI is the subset of the SQS client used by the exporter.
type sendMessageAPI interface {
	SendMessage(ctx context.Context, params *awssqs.SendMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.SendMessageOutput, error)
}

type sqsExporter struct {
	name     string
	sqs      sendMessageAPI
	queueURL string
	fifo     bool
	columns  map[string]string
	logger   *slog.Logger
}

func New(ctx context.Context, name string, cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	cfg.Logger = cfg.Logger.With("exporter", "AWS SQS")
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to SQS", slog.String("region", cfg.Region), slog.String("queue_name", cfg.QueueName), slog.String("queue_url", cfg.QueueURL))
	awsCfg, err := aws.LoadConfig(ctx, aws.Credentials{
		Region:          cfg.Region,
		AccessKeyID:     cfg.AccessKeyID,
		SecretAccessKey: cfg.SecretAccessKey,
		SessionToken:    cfg.SessionToken,
	}, cfg.Logger)
	if err != nil {
		return nil, err
	}
	client := awssqs.NewFromConfig(awsCfg)
	queueURL := cfg.QueueURL
	if queueURL == "" {
		resp, err := client.GetQueueUrl(ctx, &awssqs.GetQueueUrlInput{
			QueueName: awssdk.String(cfg.QueueName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve URL of queue %s: %w", cfg.QueueName, err)
		}
		queueURL = awssdk.ToString(resp.QueueUrl)
	}
	return newExporter(name, client, queueURL, cfg.Columns, cfg.Logger), nil
}

func newExporter(name string, client sendMessageAPI, queueURL string, columns map[string]string, logger *slog.Logger) *sqsExporter {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &sqsExporter{
		name:     name,
		sqs:      client,
		queueURL: queueURL,
		fifo:     isFIFO(queueURL),
		columns:  columns,
		logger:   logger,
	}
}

// isFIFO reports whether the queue URL refers to a FIFO queue, whose name must end with .fifo.
func isFIFO(queueURL string) bool {
	return strings.HasSuffix(strings.TrimRight(queueURL, "/"), fifoSuffix)
}

func (e *sqsExporter) Name() string {
	return e.name
}

func (e *sqsExporter) Export(ctx context.Context, data sensor.Data) error {
	body, err := json.Marshal(columnmap.Transform(e.columns, data))
	if err != nil {
		return err
	}
	attrs := make(map[string]types.MessageAttributeValue)
	if mac, ok := e.columns["mac"]; ok {
		attrs[mac] = stringAttribute(data.Addr)
	}
	if name, ok := e.columns["name"]; ok {
		attrs[name] = stringAttribute(data.Name)
	}
	input := &awssqs.SendMessageInput{
		MessageBody: awssdk.String(string(body)),
		QueueUrl:    awssdk.String(e.queueURL),
	}
	if len(attrs) > 0 {
		input.MessageAttributes = attrs
	}
	if e.fifo {
		// FIFO queues require a message group; group messages per tag so each tag's measurements stay ordered.
		input.MessageGroupId = awssdk.String(data.Addr)
		input.MessageDeduplicationId = awssdk.String(fmt.Sprintf("%s-%d", data.Addr, data.Timestamp.UnixNano()))
	}
	e.logger.LogAttrs(ctx, slog.LevelDebug, "Sending message", slog.String("queue_url", e.queueURL), slog.String("body", string(body)))
	_, err = e.sqs.SendMessage(ctx, input)
	return err
}

func (e *sqsExporter) Close() error {
	return nil
}

func stringAttribute(value string) types.MessageAttributeValue {
	return types.MessageAttributeValue{
		DataType:    awssdk.String("String"),
		StringValue: awssdk.String(value),
	}
}
