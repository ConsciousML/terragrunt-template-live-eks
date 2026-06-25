package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// unitOutput extracts a named string output from a Terragrunt stack unit.
// It fails the test immediately if the unit or the key is missing or not a string.
func unitOutput(t *testing.T, allOutputs map[string]any, unitName string, key string) string {
	t.Helper()
	unit, ok := allOutputs[unitName].(map[string]any)
	require.True(t, ok, "stack output %q missing or wrong type", unitName)
	value, ok := unit[key].(string)
	require.True(t, ok, "output %q.%s missing or wrong type", unitName, key)
	return value
}

// fetchAWSSecret retrieves the plaintext value of a Secrets Manager secret.
// It expects the secret string to be a JSON object with a "plaintext" field.
func fetchAWSSecret(t *testing.T, ctx context.Context, region string, secretName string) string {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	require.NoError(t, err, "failed to load AWS config for region %s", region)

	svc := secretsmanager.NewFromConfig(cfg)
	secret, err := svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	require.NoError(t, err, "failed to retrieve secret %q from Secrets Manager", secretName)

	var secretData struct {
		Plaintext string `json:"plaintext"`
	}
	require.NoError(t, json.Unmarshal([]byte(*secret.SecretString), &secretData), "failed to unmarshal secret JSON for %q", secretName)
	require.NotEmpty(t, secretData.Plaintext, "plaintext field is empty in secret %q", secretName)
	return secretData.Plaintext
}

// assertReachable polls url until it returns HTTP 200, retrying up to maxRetries
// times with retryInterval between attempts. It fails the test if the URL is
// never reachable within the retry budget.
func assertReachable(t *testing.T, description string, url string, maxRetries int, retryInterval time.Duration) {
	t.Helper()
	retry.DoWithRetry(t, description, maxRetries, retryInterval, func() (string, error) {
		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("GET %s: %w", url, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("GET %s: got %d, want 200", url, resp.StatusCode)
		}
		return "ready", nil
	})
}
