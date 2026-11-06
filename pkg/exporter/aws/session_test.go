//go:build aws

package aws

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsValidate(t *testing.T) {
	require.NoError(t, Credentials{}.Validate())
	require.NoError(t, Credentials{AccessKeyID: "id", SecretAccessKey: "secret"}.Validate())
	require.Error(t, Credentials{AccessKeyID: "id"}.Validate())
	require.Error(t, Credentials{SecretAccessKey: "secret"}.Validate())
}

func TestNewSession(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	t.Run("static", func(t *testing.T) {
		sess, err := NewSession(Credentials{Region: "eu-west-1", AccessKeyID: "id", SecretAccessKey: "secret"}, logger)
		require.NoError(t, err)
		assert.Equal(t, "eu-west-1", *sess.Config.Region)
		creds, err := sess.Config.Credentials.Get()
		require.NoError(t, err)
		assert.Equal(t, "id", creds.AccessKeyID)
		assert.Equal(t, "secret", creds.SecretAccessKey)
	})
	t.Run("default chain", func(t *testing.T) {
		t.Setenv("AWS_ACCESS_KEY_ID", "env-id")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "env-secret")
		t.Setenv("AWS_REGION", "us-east-2")
		sess, err := NewSession(Credentials{}, logger)
		require.NoError(t, err)
		assert.Equal(t, "us-east-2", *sess.Config.Region)
		creds, err := sess.Config.Credentials.Get()
		require.NoError(t, err)
		assert.Equal(t, "env-id", creds.AccessKeyID)
	})
}
