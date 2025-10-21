package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

func newLocalstackClient(t *testing.T) *ssm.Client {
	t.Helper()

	endpoint := localstackEndpoint()

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "1")
	t.Setenv("AWS_ENDPOINT_URL", endpoint)
	t.Setenv("AWS_ENDPOINT_URL_SSM", endpoint)
	t.Setenv("AWS_SSM_ENDPOINT", endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	cfg, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err, "load aws config")

	client := ssm.NewFromConfig(cfg)

	ctxHealth, cancelHealth := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelHealth()

	_, err = client.DescribeParameters(ctxHealth, &ssm.DescribeParametersInput{
		MaxResults: aws.Int32(1),
	})
	require.NoErrorf(t, err, "describe parameters against localstack (%s)", endpoint)

	return client
}

func uniqueParameterName(prefix string) string {
	return fmt.Sprintf("sfx-test-%s-%d", prefix, time.Now().UnixNano())
}

func createParameter(t *testing.T, client *ssm.Client, name string, value string, paramType ssmtypes.ParameterType) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(name),
		Value: aws.String(value),
		Type:  paramType,
	})
	require.NoErrorf(t, err, "put parameter %s", name)

	t.Cleanup(func() {
		ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelDelete()
		_, _ = client.DeleteParameter(ctxDelete, &ssm.DeleteParameterInput{
			Name: aws.String(name),
		})
	})
}

func getParameterValue(t *testing.T, client *ssm.Client, name string, withDecryption bool) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(withDecryption),
	})
	require.NoErrorf(t, err, "get parameter %s (with_decryption=%t)", name, withDecryption)
	require.NotNil(t, resp.Parameter, "parameter response")
	require.NotNil(t, resp.Parameter.Value, "parameter value")

	return aws.ToString(resp.Parameter.Value)
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

func TestHandleRequiresParameterName(t *testing.T) {
	t.Parallel()

	_, err := handle(provider.Request{Ref: ""})
	if assert.Error(t, err) {
		assert.EqualError(t, err, "ref must include the parameter name")
	}
}

func TestHandleReturnsParameterValue(t *testing.T) {
	client := newLocalstackClient(t)
	paramName := uniqueParameterName("string")
	createParameter(t, client, paramName, "plain-value", ssmtypes.ParameterTypeString)

	resp, err := handle(provider.Request{Ref: paramName})
	require.NoError(t, err)
	assert.Equal(t, "plain-value", string(resp.Value))
}

func TestHandleDecryptsSecureStringByDefault(t *testing.T) {
	client := newLocalstackClient(t)
	paramName := uniqueParameterName("secure-default")
	createParameter(t, client, paramName, "top-secret", ssmtypes.ParameterTypeSecureString)

	resp, err := handle(provider.Request{Ref: paramName})
	require.NoError(t, err)
	assert.Equal(t, "top-secret", string(resp.Value))
}

func TestHandleHonorsWithDecryptionOption(t *testing.T) {
	client := newLocalstackClient(t)
	paramName := uniqueParameterName("secure-option")
	createParameter(t, client, paramName, "option-secret", ssmtypes.ParameterTypeSecureString)

	encryptedValue := getParameterValue(t, client, paramName, false)
	require.NotEqual(t, "option-secret", encryptedValue, "expected encrypted value to differ from plaintext")

	resp, err := handle(provider.Request{
		Ref:     paramName,
		Options: []byte("with_decryption: false\n"),
	})
	require.NoError(t, err)
	assert.Equal(t, encryptedValue, string(resp.Value))
}
