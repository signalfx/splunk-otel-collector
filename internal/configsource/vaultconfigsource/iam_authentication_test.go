// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vaultconfigsource

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

type capturedIAMLogin struct {
	decodeErr   error
	loginData   map[string]any
	requestPath string
}

func TestIAMAuthenticationToken(t *testing.T) {
	accessKeyID := "AKIDEXAMPLE"
	secretAccessKey := "fake-secret-access-key" // #nosec G101 -- deliberately fake credential for request-signing test
	securityToken := "fake-security-token"
	headerValue := "vault.example.com"
	mount := "custom-aws"
	role := "test-role"

	client, captured := newTestVaultClient(t)

	authentication := IAMAuthentication{
		AWSAccessKeyID:     &accessKeyID,
		AWSSecretAccessKey: &secretAccessKey,
		AWSSecurityToken:   &securityToken,
		HeaderValue:        &headerValue,
		Mount:              &mount,
		Role:               &role,
	}
	token, err := authentication.Token(client)
	require.NoError(t, err)
	require.Equal(t, "test-vault-token", token)
	require.NoError(t, captured.decodeErr)
	require.Equal(t, "/v1/auth/custom-aws/login", captured.requestPath)
	require.Equal(t, role, captured.loginData["role"])
	require.Equal(t, http.MethodPost, captured.loginData["iam_http_request_method"])

	headers := decodeIAMRequestHeaders(t, captured.loginData)
	require.Equal(t, headerValue, headers.Get("X-Vault-AWS-IAM-Server-ID"))
	require.Equal(t, securityToken, headers.Get("X-Amz-Security-Token"))
	require.Contains(t, headers.Get("Authorization"), "Credential="+accessKeyID+"/")
	require.Contains(t, headers.Get("Authorization"), "/us-east-1/sts/aws4_request")

	encodedBody, ok := captured.loginData["iam_request_body"].(string)
	require.True(t, ok)
	requestBody, err := base64.StdEncoding.DecodeString(encodedBody)
	require.NoError(t, err)
	require.Contains(t, string(requestBody), "Action=GetCallerIdentity")
}

func TestIAMAuthenticationTokenEnvironmentCredentials(t *testing.T) {
	accessKeyID := "ENVAKIDEXAMPLE"
	secretAccessKey := "fake-environment-secret-access-key" // #nosec G101 -- deliberately fake credential for request-signing test
	securityToken := "fake-environment-security-token"
	t.Setenv("AWS_ACCESS_KEY_ID", accessKeyID)
	t.Setenv("AWS_SECRET_ACCESS_KEY", secretAccessKey)
	t.Setenv("AWS_SESSION_TOKEN", securityToken)
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	client, captured := newTestVaultClient(t)
	token, err := (&IAMAuthentication{}).Token(client)
	require.NoError(t, err)
	require.Equal(t, "test-vault-token", token)
	require.NoError(t, captured.decodeErr)
	require.Equal(t, "/v1/auth/aws/login", captured.requestPath)
	require.Empty(t, captured.loginData["role"])

	headers := decodeIAMRequestHeaders(t, captured.loginData)
	require.Equal(t, securityToken, headers.Get("X-Amz-Security-Token"))
	require.Contains(t, headers.Get("Authorization"), "Credential="+accessKeyID+"/")
	require.Contains(t, headers.Get("Authorization"), "/us-east-1/sts/aws4_request")
}

func newTestVaultClient(t *testing.T) (*api.Client, *capturedIAMLogin) {
	t.Helper()
	captured := &capturedIAMLogin{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.requestPath = r.URL.Path
		captured.decodeErr = json.NewDecoder(r.Body).Decode(&captured.loginData)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"auth":{"client_token":"test-vault-token"}}`))
	}))
	t.Cleanup(server.Close)

	config := api.DefaultConfig()
	config.Address = server.URL
	client, err := api.NewClient(config)
	require.NoError(t, err)
	return client, captured
}

func decodeIAMRequestHeaders(t *testing.T, loginData map[string]any) http.Header {
	t.Helper()
	encodedHeaders, ok := loginData["iam_request_headers"].(string)
	require.True(t, ok)
	headerJSON, err := base64.StdEncoding.DecodeString(encodedHeaders)
	require.NoError(t, err)
	var headers http.Header
	require.NoError(t, json.Unmarshal(headerJSON, &headers))
	return headers
}
