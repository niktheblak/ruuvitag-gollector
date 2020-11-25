package sqs

type Config struct {
	QueueName       string
	QueueURL        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}
