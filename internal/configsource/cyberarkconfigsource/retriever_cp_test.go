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

package cyberarkconfigsource

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// joinFields builds a well-formed CLIPasswordSDK output line from ordered values.
func joinFields(values ...string) []byte {
	return []byte(strings.Join(values, outputDelimiter) + "\n")
}

func TestCPRetriever_Retrieve(t *testing.T) {
	// Ordered per outputFields: Password, UserName, Address, Database, Port, Name, Safe, Folder.
	out := joinFields("s3cr3t", "svc", "db.example.com", "app", "5432", "prod-db", "DBSecrets", "Root")

	r := &cpRetriever{
		binaryPath: defaultBinaryPath,
		appID:      "collector-app",
		safe:       "DBSecrets",
		object:     "prod-db",
		runner: func(context.Context, string, ...string) ([]byte, error) {
			return out, nil
		},
	}

	fields, err := r.retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "s3cr3t", fields["Password"])
	assert.Equal(t, "svc", fields["UserName"])
	assert.Equal(t, "db.example.com", fields["Address"])
	assert.Equal(t, "5432", fields["Port"])
	assert.Equal(t, "Root", fields["Folder"])
}

func TestCPRetriever_ExecError(t *testing.T) {
	r := &cpRetriever{
		binaryPath: defaultBinaryPath,
		runner: func(context.Context, string, ...string) ([]byte, error) {
			return nil, errors.New("exit status 1: object not found")
		},
	}

	fields, err := r.retrieve(context.Background())
	require.Error(t, err)
	assert.Nil(t, fields)
	assert.IsType(t, &errCLIExec{}, err)
}

func TestParseOutput_WrongFieldCount(t *testing.T) {
	// Only two fields, but outputFields expects more.
	fields, err := parseOutput(joinFields("s3cr3t", "svc"))
	require.Error(t, err)
	assert.Nil(t, fields)
	assert.IsType(t, &errParseOutput{}, err)
}

func TestBuildArgs_FolderDefaultsToRoot(t *testing.T) {
	r := &cpRetriever{
		binaryPath: defaultBinaryPath,
		appID:      "collector-app",
		safe:       "DBSecrets",
		folder:     "", // empty -> Root
		object:     "prod-db",
	}

	args := r.buildArgs()
	joined := strings.Join(args, " ")

	assert.Equal(t, "GetPassword", args[0])
	assert.Contains(t, joined, "AppDescs.AppID=collector-app")
	assert.Contains(t, joined, "Query=Safe=DBSecrets;Folder=Root;Object=prod-db")
	assert.Contains(t, joined, "-d "+outputDelimiter)
	// The -o list must request every configured attribute.
	for _, f := range outputFields {
		assert.Contains(t, joined, f.attr)
	}
}

func TestBuildArgs_ExplicitFolder(t *testing.T) {
	r := &cpRetriever{
		binaryPath: defaultBinaryPath,
		appID:      "collector-app",
		safe:       "DBSecrets",
		folder:     "Nested",
		object:     "prod-db",
	}

	joined := strings.Join(r.buildArgs(), " ")
	assert.Contains(t, joined, "Query=Safe=DBSecrets;Folder=Nested;Object=prod-db")
}
