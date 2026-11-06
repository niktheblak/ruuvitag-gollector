//go:build aws

package aws

import (
	"context"
	"fmt"
	"log/slog"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// Credentials contains the common AWS connection settings shared by the AWS exporters.
//
// When AccessKeyID and SecretAccessKey are both set, static credentials are used.
// Otherwise the default AWS credential chain (environment, shared config, instance role) is used.
type Credentials struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func (c Credentials) static() bool {
	return c.AccessKeyID != "" && c.SecretAccessKey != ""
}

// Validate checks that the credentials are either fully specified or not specified at all.
func (c Credentials) Validate() error {
	if (c.AccessKeyID == "") != (c.SecretAccessKey == "") {
		return fmt.Errorf("AWS access key ID and secret access key must both be specified or both be empty")
	}
	return nil
}

// LoadConfig creates an AWS SDK configuration from the given credentials.
func LoadConfig(ctx context.Context, creds Credentials, logger *slog.Logger) (awssdk.Config, error) {
	if err := creds.Validate(); err != nil {
		return awssdk.Config{}, err
	}
	var opts []func(*config.LoadOptions) error
	if creds.Region != "" {
		opts = append(opts, config.WithRegion(creds.Region))
	}
	if creds.static() {
		logger.Debug("Using static AWS credentials")
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)))
	} else {
		logger.Debug("Using default AWS credential chain")
	}
	return config.LoadDefaultConfig(ctx, opts...)
}
