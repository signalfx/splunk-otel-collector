// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension"
)

const ContextKey = splunkauthextension.ContextKey

type HecTokenConfig = splunkauthextension.HecTokenConfig

func readTokens(f feed) []HecTokenConfig {
	tokens := make([]HecTokenConfig, len(f.Entry))
	for i := range f.Entry {
		var token configopaque.String
		var defaultIndex string
		var indexes []string
		for _, key := range f.Entry[i].Content.Dict.Key {
			if key.Name == "token" {
				token = configopaque.String(key.Text)
			}
			if key.Name == "index" {
				defaultIndex = key.Text
			}
			if key.Name == "indexes" {
				indexes = append(indexes, key.List.Item...)
			}
		}
		tokens[i] = splunkauthextension.HecTokenConfig{
			Token:          token,
			DefaultIndex:   defaultIndex,
			AllowedIndexes: indexes,
		}
	}
	return tokens
}

func New(ctx context.Context, settings component.TelemetrySettings, serverURI, sessionKey string) (extension.Extension, error) {
	req, err := http.NewRequest(http.MethodGet, serverURI+"/servicesNS/-/-/data/inputs/http", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Splunk "+sessionKey)
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	f := feed{}
	err = xml.Unmarshal(body, &f)
	if err != nil {
		return nil, err
	}

	tokens := readTokens(f)

	sae := splunkauthextension.NewFactory()
	authConfig := sae.CreateDefaultConfig().(*splunkauthextension.Config)
	authConfig.Tokens = tokens

	return sae.Create(ctx, extension.Settings{
		ID:                component.MustNewID("splunkauth"),
		TelemetrySettings: settings,
	}, authConfig)
}

// feed is generated from the atom feed XML response of the Splunk instance, per
// https://help.splunk.com/en/splunk-enterprise/leverage-rest-apis/rest-api-user-manual/9.3/rest-api-user-manual/basic-concepts-about-the-splunk-platform-rest-api
type feed struct {
	XMLName    xml.Name `xml:"feed"`
	Text       string   `xml:",chardata"`
	Xmlns      string   `xml:"xmlns,attr"`
	S          string   `xml:"s,attr"`
	Opensearch string   `xml:"opensearch,attr"`
	Title      string   `xml:"title"`
	ID         string   `xml:"id"`
	Updated    string   `xml:"updated"`
	Generator  struct {
		Text    string `xml:",chardata"`
		Build   string `xml:"build,attr"`
		Version string `xml:"version,attr"`
	} `xml:"generator"`
	Author struct {
		Text string `xml:",chardata"`
		Name string `xml:"name"`
	} `xml:"author"`
	Link []struct {
		Text string `xml:",chardata"`
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
	TotalResults string `xml:"totalResults"`
	ItemsPerPage string `xml:"itemsPerPage"`
	StartIndex   string `xml:"startIndex"`
	Messages     string `xml:"messages"`
	Entry        []struct {
		Text    string `xml:",chardata"`
		Title   string `xml:"title"`
		ID      string `xml:"id"`
		Updated string `xml:"updated"`
		Link    []struct {
			Text string `xml:",chardata"`
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
		} `xml:"link"`
		Author struct {
			Text string `xml:",chardata"`
			Name string `xml:"name"`
		} `xml:"author"`
		Content struct {
			Text string `xml:",chardata"`
			Type string `xml:"type,attr"`
			Dict struct {
				Text string `xml:",chardata"`
				Key  []struct {
					List struct {
						Text string   `xml:",chardata"`
						Item []string `xml:"item"`
					} `xml:"list"`
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
					Dict struct {
						Text string `xml:",chardata"`
						Key  []struct {
							Text string `xml:",chardata"`
							Name string `xml:"name,attr"`
							Dict struct {
								Text string `xml:",chardata"`
								Key  []struct {
									Text string `xml:",chardata"`
									Name string `xml:"name,attr"`
									List struct {
										Text string `xml:",chardata"`
										Item string `xml:"item"`
									} `xml:"list"`
								} `xml:"key"`
							} `xml:"dict"`
						} `xml:"key"`
					} `xml:"dict"`
				} `xml:"key"`
			} `xml:"dict"`
		} `xml:"content"`
	} `xml:"entry"`
}
