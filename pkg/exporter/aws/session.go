//go:build aws

package aws

import (
	"fmt"
	"log/slog"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
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

// NewSession creates an AWS session from the given credentials.
func NewSession(creds Credentials, logger *slog.Logger) (*session.Session, error) {
	if err := creds.Validate(); err != nil {
		return nil, err
	}
	cfg := awssdk.NewConfig()
	if creds.Region != "" {
		cfg = cfg.WithRegion(creds.Region)
	}
	if creds.static() {
		logger.Debug("Using static AWS credentials")
		cfg = cfg.WithCredentials(credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken))
	} else {
		logger.Debug("Using default AWS credential chain")
	}
	return session.NewSessionWithOptions(session.Options{
		Config:            *cfg,
		SharedConfigState: session.SharedConfigEnable,
	})
}
