// Copyright  Splunk, Inc.
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

package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestWarningProviderRetrieve(t *testing.T) {
	t.Setenv("TEST", "test-value")
	wfp := &warningProviderFactory{
		ProviderFactory: envprovider.NewFactory(),
		warnings:        map[string]string{"TEST": "This is a test warning"},
	}

	logCore, logObserver := observer.New(zap.WarnLevel)
	logger := zap.New(logCore)
	wp := wfp.Create(confmap.ProviderSettings{Logger: logger})

	retrieved, err := wp.Retrieve(context.Background(), "env:TEST", nil)
	require.NoError(t, err)
	rawVal, rErr := retrieved.AsRaw()
	require.NoError(t, rErr)
	assert.Equal(t, "test-value", rawVal.(string))

	assert.Equal(t, 1, logObserver.Len())
	assert.Equal(t, "This is a test warning", logObserver.All()[0].Message)
}
