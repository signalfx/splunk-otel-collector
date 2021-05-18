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

package templateconfigsource

import (
	"bytes"
	"context"
	"text/template"

	"go.opentelemetry.io/collector/config/experimental/configsource"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// templateSession implements the configsource.Session interface.
type templateSession struct{}

var _ configsource.Session = (*templateSession)(nil)

func (ts *templateSession) Retrieve(_ context.Context, selector string, params interface{}) (configsource.Retrieved, error) {
	tmpl, err := template.ParseFiles(selector)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, err
	}

	return configprovider.NewRetrieved(buf.Bytes(), configprovider.WatcherNotSupported), nil
}

func (ts *templateSession) RetrieveEnd(context.Context) error {
	return nil
}

func (ts *templateSession) Close(context.Context) error {
	return nil
}

func newSession() (*templateSession, error) {
	return &templateSession{}, nil
}
