// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configmapprovider"
)

func TestReplaceDollarDollar(t *testing.T) {
	pp := &converterProvider{
		wrapped:     configmapprovider.NewFile("testdata/dollar-dollar.yaml"),
		cfgMapFuncs: []CfgMapFunc{ReplaceDollarDollar},
	}
	r, err := pp.Retrieve(context.Background(), nil)
	require.NoError(t, err)
	v, err := r.Get(context.Background())
	require.NoError(t, err)
	endpt := v.Get("exporters::otlp::endpoint")
	assert.Equal(t, "localhost:${env:OTLP_PORT}", endpt)
	insecure := v.Get("exporters::otlp::insecure")
	assert.Equal(t, "$OTLP_INSECURE", insecure)
	pwd := v.Get("receivers::redis::password")
	assert.Equal(t, "$$ecret", pwd)
	host := v.Get("receivers::redis::host")
	assert.Equal(t, "ho$$tname:${env:PORT}", host)
}

func TestRegexReplace(t *testing.T) {
	assert.Equal(t, "${foo/bar:PORT}", testReplace("$${foo/bar:PORT}"))
	assert.Equal(t, "$${PORT}", testReplace("$${PORT}"))
}

func testReplace(s string) string {
	return replaceDollarDollar(dollarDollarRegex(), s)
}
