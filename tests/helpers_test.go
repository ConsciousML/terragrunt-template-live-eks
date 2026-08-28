package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

// postLoginJSON POSTs a JSON-encoded request body to url and decodes the JSON response into out.
// It fails the test on a marshal/request/decode error or a non-200 status.
func postLoginJSON(t *testing.T, url string, reqBody any, out any) {
	t.Helper()

	body, err := json.Marshal(reqBody)
	require.NoError(t, err, "failed to marshal request body for %s", url)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	require.NoError(t, err, "POST %s failed", url)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "login to %s returned unexpected status: got %d, want 200", url, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(out), "failed to decode login response from %s", url)
}
