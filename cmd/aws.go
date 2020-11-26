// +build aws

package cmd

import (
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/dynamodb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/aws/sqs"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().String("aws.region", "us-east-2", "AWS region")
	rootCmd.PersistentFlags().String("aws.access_key_id", "", "AWS access key ID")
	rootCmd.PersistentFlags().String("aws.secret_access_key", "", "AWS secret access key")
	rootCmd.PersistentFlags().String("aws.session_token", "", "AWS session token")
	rootCmd.PersistentFlags().Bool("aws.dynamodb.enabled", false, "Store measurements to AWS DynamoDB")
	rootCmd.PersistentFlags().String("aws.dynamodb.table", "", "AWS DynamoDB table name")
	rootCmd.PersistentFlags().Bool("aws.sqs.enabled", false, "Send measurements to AWS SQS")
	rootCmd.PersistentFlags().String("aws.sqs.queue.name", "", "AWS SQS queue name")
	rootCmd.PersistentFlags().String("aws.sqs.queue.url", "", "AWS SQS queue URL")
}

func addDynamoDBExporter(exporters *[]exporter.Exporter) error {
	table := viper.GetString("aws.dynamodb.table")
	if table == "" {
		return fmt.Errorf("DynamoDB table name must be specified")
	}
	exp, err := dynamodb.New(dynamodb.Config{
		Table:           table,
		Region:          viper.GetString("aws.region"),
		AccessKeyID:     viper.GetString("aws.access_key_id"),
		SecretAccessKey: viper.GetString("aws.secret_access_key"),
		SessionToken:    viper.GetString("aws.session_token"),
	})
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}

func addSQSExporter(exporters *[]exporter.Exporter) error {
	queueName := viper.GetString("aws.sqs.queue.name")
	queueURL := viper.GetString("aws.sqs.queue.url")
	if queueName == "" && queueURL == "" {
		return fmt.Errorf("AWS SQS queue name or queue URL must be specified")
	}
	exp, err := sqs.New(sqs.Config{
		QueueName:       queueName,
		QueueURL:        queueURL,
		Region:          viper.GetString("aws.region"),
		AccessKeyID:     viper.GetString("aws.access_key_id"),
		SecretAccessKey: viper.GetString("aws.secret_access_key"),
		SessionToken:    viper.GetString("aws.session_token"),
	})
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
