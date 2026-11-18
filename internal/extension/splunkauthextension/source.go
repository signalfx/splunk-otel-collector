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

package splunkauthextension

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/config/configopaque"
)

func readTokens(f feed) []HecTokenConfig {
	tokens := make([]HecTokenConfig, len(f.Entry))
	for i := range f.Entry {
		var token configopaque.String
		var disabled bool
		var defaultIndex string
		var indexes []string
		for _, key := range f.Entry[i].Content.Dict.Key {
			if key.Name == "disabled" {
				// The server normalizes boolean values to "1" and "0"; fail closed on
				// anything else so an unexpected value never enables a token.
				disabled = strings.TrimSpace(key.Text) != "0"
			}
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
		tokens[i] = HecTokenConfig{
			Token:          token,
			Disabled:       disabled,
			DefaultIndex:   defaultIndex,
			AllowedIndexes: indexes,
		}
	}
	return tokens
}

func newTokenSource(ctx context.Context, cfg *Config) (tokenSource, error) {
	tlsConfig, err := cfg.TLS.LoadTLSConfig(ctx)
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	return &splunkTokenSource{
		serverURI:  cfg.Endpoint,
		sessionKey: string(cfg.SessionKey),
		client:     &http.Client{Transport: transport},
	}, nil
}

type splunkTokenSource struct {
	client     *http.Client
	serverURI  string
	sessionKey string
}

func (s *splunkTokenSource) Load(ctx context.Context) ([]HecTokenConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(s.serverURI, "/")+"/servicesNS/-/-/data/inputs/http", http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Splunk "+s.sessionKey)
	var tokens []HecTokenConfig
	// Use Splunk's default page size, advancing the offset by the number of
	// entries actually returned.
	// https://help.splunk.com/en/splunk-enterprise/rest-api-reference/9.2/introduction/using-the-rest-api-reference
	for offset := 0; ; {
		query := req.URL.Query()
		query.Set("offset", strconv.Itoa(offset))
		req.URL.RawQuery = query.Encode()
		f, err := readTokenPage(s.client, req)
		if err != nil {
			return nil, fmt.Errorf("retrieve HEC tokens at offset %d: %w", offset, err)
		}
		if f.TotalResults == nil {
			return nil, fmt.Errorf("missing HEC token totalResults at offset %d", offset)
		}
		if f.StartIndex != offset || *f.TotalResults < 0 {
			return nil, fmt.Errorf("invalid HEC token pagination at offset %d", offset)
		}
		tokens = append(tokens, readTokens(f)...)
		offset += len(f.Entry)
		if offset >= *f.TotalResults {
			break
		}
		if len(f.Entry) == 0 {
			return nil, fmt.Errorf("empty HEC token page at offset %d before totalResults %d", offset, *f.TotalResults)
		}
	}

	return tokens, nil
}

func readTokenPage(client *http.Client, req *http.Request) (feed, error) {
	resp, err := client.Do(req)
	if err != nil {
		return feed{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return feed{}, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	var f feed
	err = xml.NewDecoder(resp.Body).Decode(&f)
	return f, err
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
	TotalResults *int   `xml:"totalResults"`
	ItemsPerPage string `xml:"itemsPerPage"`
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
	StartIndex int `xml:"startIndex"`
}
