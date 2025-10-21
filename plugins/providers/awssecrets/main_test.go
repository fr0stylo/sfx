package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fr0stylo/sfx/provider"
)

const defaultLocalstackEndpoint = "http://localhost:4566"

func localstackEndpoint() string {
	if endpoint := os.Getenv("SFX_LOCALSTACK_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return defaultLocalstackEndpoint
}

func newLocalstackClient(t *testing.T) *secretsmanager.Client {
	t.Helper()

	endpoint := localstackEndpoint()

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "1")
	t.Setenv("AWS_ENDPOINT_URL", endpoint)
	t.Setenv("AWS_ENDPOINT_URL_SECRETSMANAGER", endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err, "load aws config")

	client := secretsmanager.NewFromConfig(cfg)

	ctxHealth, cancelHealth := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelHealth()

	_, err = client.ListSecrets(ctxHealth, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(1),
	})
	require.NoErrorf(t, err, "list secrets against localstack (%s)", endpoint)

	return client
}

func uniqueSecretName(prefix string) string {
	return fmt.Sprintf("sfx-test-%s-%d", prefix, time.Now().UnixNano())
}

func createSecretString(t *testing.T, client *secretsmanager.Client, name string, value string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(value),
	})
	require.NoErrorf(t, err, "create secret %s", name)

	t.Cleanup(func() {
		ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelDelete()
		_, _ = client.DeleteSecret(ctxDelete, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(name),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	return aws.ToString(out.VersionId)
}

func putSecretString(t *testing.T, client *secretsmanager.Client, name string, value string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(name),
		SecretString: aws.String(value),
	})
	require.NoErrorf(t, err, "put secret value %s", name)

	return aws.ToString(out.VersionId)
}

func createSecretBinary(t *testing.T, client *secretsmanager.Client, name string, value []byte) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretBinary: value,
	})
	require.NoErrorf(t, err, "create binary secret %s", name)

	t.Cleanup(func() {
		ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelDelete()
		_, _ = client.DeleteSecret(ctxDelete, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(name),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		})
	})

	return aws.ToString(out.VersionId)
}

func TestParseRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ref          string
		secretID     string
		versionID    string
		versionStage string
	}{
		{
			name:         "simple secret",
			ref:          "my-secret",
			secretID:     "my-secret",
			versionID:    "",
			versionStage: "",
		},
		{
			name:         "version via version prefix",
			ref:          "my-secret#version:abc123",
			secretID:     "my-secret",
			versionID:    "abc123",
			versionStage: "",
		},
		{
			name:         "version via version_id prefix",
			ref:          "  my-secret  #version_id:abc456  ",
			secretID:     "my-secret",
			versionID:    "abc456",
			versionStage: "",
		},
		{
			name:         "stage via stage prefix",
			ref:          "my-secret#stage:AWSPREVIOUS",
			secretID:     "my-secret",
			versionID:    "",
			versionStage: "AWSPREVIOUS",
		},
		{
			name:         "stage via default suffix",
			ref:          "my-secret#AWSCURRENT",
			secretID:     "my-secret",
			versionID:    "",
			versionStage: "AWSCURRENT",
		},
		{
			name:         "empty ref",
			ref:          "  ",
			secretID:     "",
			versionID:    "",
			versionStage: "",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			secretID, versionID, versionStage := parseRef(tt.ref)
			assert.Equal(t, tt.secretID, secretID, "secret id")
			assert.Equal(t, tt.versionID, versionID, "version id")
			assert.Equal(t, tt.versionStage, versionStage, "version stage")
		})
	}
}

func TestResolveTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		given    time.Duration
		fallback time.Duration
		want     time.Duration
	}{
		{
			name:     "uses fallback when zero",
			given:    0,
			fallback: 30 * time.Second,
			want:     30 * time.Second,
		},
		{
			name:     "uses fallback when negative",
			given:    -5 * time.Second,
			fallback: 10 * time.Second,
			want:     10 * time.Second,
		},
		{
			name:     "keeps positive timeout",
			given:    15 * time.Second,
			fallback: 10 * time.Second,
			want:     15 * time.Second,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := resolveTimeout(tt.given, tt.fallback)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandleRequiresSecretID(t *testing.T) {
	t.Parallel()

	_, err := handle(provider.Request{Ref: ""})
	if assert.Error(t, err) {
		assert.EqualError(t, err, "ref must include the secret identifier")
	}
}

func TestHandleReturnsSecretString(t *testing.T) {
	client := newLocalstackClient(t)
	secretName := uniqueSecretName("string")
	versionID := createSecretString(t, client, secretName, "super-secret")

	resp, err := handle(provider.Request{Ref: fmt.Sprintf("%s#version:%s", secretName, versionID)})
	require.NoError(t, err)
	assert.Equal(t, "super-secret", string(resp.Value))
}

func TestHandlePrefersOptionVersionStage(t *testing.T) {
	client := newLocalstackClient(t)
	secretName := uniqueSecretName("stage")

	prevVersion := createSecretString(t, client, secretName, "stage-secret")
	_ = putSecretString(t, client, secretName, "current-secret")

	require.NotEmpty(t, prevVersion, "previous version id must not be empty")

	resp, err := handle(provider.Request{
		Ref:     secretName + "#stage:IGNORED",
		Options: []byte("version_stage: AWSPREVIOUS\n"),
	})
	require.NoError(t, err)
	assert.Equal(t, "stage-secret", string(resp.Value))
}

func TestHandleReturnsBinary(t *testing.T) {
	client := newLocalstackClient(t)
	secretName := uniqueSecretName("binary")
	createSecretBinary(t, client, secretName, []byte("test"))

	resp, err := handle(provider.Request{Ref: secretName})
	require.NoError(t, err)
	assert.Equal(t, "test", string(resp.Value))
}
