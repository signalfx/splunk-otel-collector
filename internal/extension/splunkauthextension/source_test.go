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
	_ "embed"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensionauth"

	"github.com/signalfx/splunk-otel-collector/internal/auth/authtest"
)

func newTestAuth(ctx context.Context, settings component.TelemetrySettings, serverURI, sessionKey string) (extension.Extension, error) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Endpoint, cfg.SessionKey = serverURI, configopaque.String(sessionKey)
	e, err := factory.Create(ctx, extension.Settings{ID: component.MustNewID("splunk_auth"), TelemetrySettings: settings}, cfg)
	if err != nil {
		return nil, err
	}
	if errStart := e.Start(ctx, componenttest.NewNopHost()); errStart != nil {
		return nil, errStart
	}
	return e, nil
}

func TestAuth(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	authtest.SetupAuth(t, listener)
	e, err := newTestAuth(context.Background(), componenttest.NewNopTelemetrySettings(), "http://"+listener.Addr().String(), "foo")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, e.Shutdown(context.Background())) })
	require.NotNil(t, e)
	se := e.(extensionauth.Server)
	ctx, err := se.Authenticate(context.Background(), map[string][]string{
		"foo":           {"bar"},
		"Authorization": {"Splunk 00000000-0000-0000-0000-0000000000000"},
	})
	require.NoError(t, err)
	require.NotNil(t, ctx)
	cfg := ctx.Value(ContextKey).(HecTokenConfig)
	require.Equal(t, []string{"main", "foo"}, cfg.AllowedIndexes)
	_, err = se.Authenticate(context.Background(), map[string][]string{
		"foo":           {"bar"},
		"Authorization": {"Splunk 00000000-0000-0000-0000-111111111111"},
	})
	require.Error(t, err)
}

func TestReadTokensDisabled(t *testing.T) {
	for _, tc := range []struct {
		name     string
		field    string
		disabled bool
	}{
		{name: "enabled", field: `<s:key name="disabled">0</s:key>`},
		{name: "disabled", field: `<s:key name="disabled">1</s:key>`, disabled: true},
		{name: "unexpected value", field: `<s:key name="disabled">true</s:key>`, disabled: true},
		{name: "empty value", field: `<s:key name="disabled"></s:key>`, disabled: true},
		{name: "missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.Replace(tokenPageXML(1, 0, 1), "<s:dict>", "<s:dict>"+tc.field, 1)
			var f feed
			require.NoError(t, xml.Unmarshal([]byte(body), &f))
			require.Equal(t, []HecTokenConfig{{
				Token:          configopaque.String("token-0"),
				Disabled:       tc.disabled,
				DefaultIndex:   "index-0",
				AllowedIndexes: []string{"index-0", "main"},
			}}, readTokens(f))
		})
	}
}

func TestAuthDisabledToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body := strings.Replace(tokenPageXML(2, 0, 2), "<s:dict>", `<s:dict><s:key name="disabled">1</s:key>`, 1)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	e, err := newTestAuth(t.Context(), componenttest.NewNopTelemetrySettings(), server.URL, "session-key")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, e.Shutdown(context.Background())) })
	se := e.(extensionauth.Server)
	ctx, err := se.Authenticate(t.Context(), map[string][]string{
		"Authorization": {"Splunk token-0"},
	})
	require.EqualError(t, err, "HEC token is disabled")
	require.Nil(t, ctx.Value(ContextKey))

	ctx, err = se.Authenticate(t.Context(), map[string][]string{
		"Authorization": {"Splunk token-1"},
	})
	require.NoError(t, err)
	require.Equal(t, "index-1", ctx.Value(ContextKey).(HecTokenConfig).DefaultIndex)
}

func TestAuthPagination(t *testing.T) {
	for _, tc := range []struct {
		name     string
		total    int
		pageSize int
	}{
		{name: "empty", total: 0, pageSize: 30},
		{name: "single page", total: 2, pageSize: 30},
		{name: "full final page", total: 60, pageSize: 30},
		{name: "partial final page", total: 65, pageSize: 30},
		{name: "server caps page size", total: 25, pageSize: 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var offsets []int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/servicesNS/-/-/data/inputs/http", r.URL.Path)
				assert.Equal(t, "Splunk session-key", r.Header.Get("Authorization"))
				assert.False(t, r.URL.Query().Has("count"))
				offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
				if !assert.NoError(t, err) {
					http.Error(w, "invalid offset", http.StatusBadRequest)
					return
				}
				offsets = append(offsets, offset)
				_, _ = fmt.Fprint(w, tokenPageXML(tc.total, offset, min(tc.pageSize, tc.total-offset)))
			}))
			defer server.Close()

			e, err := newTestAuth(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "session-key")
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, e.Shutdown(context.Background())) })
			var expectedOffsets []int
			for offset := 0; ; offset += tc.pageSize {
				expectedOffsets = append(expectedOffsets, offset)
				if offset+tc.pageSize >= tc.total {
					break
				}
			}
			require.Equal(t, expectedOffsets, offsets)
			for i := 0; i < tc.total; i++ {
				ctx, err := e.(extensionauth.Server).Authenticate(context.Background(), map[string][]string{
					"Authorization": {fmt.Sprintf("Splunk token-%d", i)},
				})
				require.NoError(t, err)
				cfg := ctx.Value(ContextKey).(HecTokenConfig)
				require.Equal(t, fmt.Sprintf("index-%d", i), cfg.DefaultIndex)
				require.Equal(t, []string{fmt.Sprintf("index-%d", i), "main"}, cfg.AllowedIndexes)
			}
		})
	}
}

func TestAuthPaginationFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		body   string
		want   string
		status int
	}{
		{name: "HTTP error", status: http.StatusForbidden, body: "denied", want: "403 Forbidden"},
		{name: "malformed XML", status: http.StatusOK, body: "<feed>", want: "XML syntax error"},
		{name: "empty page", status: http.StatusOK, body: tokenPageXML(2, 1, 0), want: "empty HEC token page"},
		{name: "repeated page", status: http.StatusOK, body: tokenPageXML(2, 0, 1), want: "invalid HEC token pagination"},
		{name: "invalid total", status: http.StatusOK, body: strings.ReplaceAll(tokenPageXML(2, 1, 1), "<opensearch:totalResults>2", "<opensearch:totalResults>invalid"), want: "invalid syntax"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if requests == 1 {
					_, _ = fmt.Fprint(w, tokenPageXML(2, 0, 1))
					return
				}
				assert.Equal(t, "1", r.URL.Query().Get("offset"))
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			e, err := newTestAuth(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "foo")
			require.ErrorContains(t, err, tc.want)
			require.ErrorContains(t, err, "offset 1")
			require.Nil(t, e)
			require.Equal(t, 2, requests)
		})
	}
}

func TestAuthInvalidTotalResults(t *testing.T) {
	for _, tc := range []struct {
		name  string
		total string
		want  string
	}{
		{name: "missing", total: "", want: "missing HEC token totalResults"},
		{name: "negative", total: "<opensearch:totalResults>-1</opensearch:totalResults>", want: "invalid HEC token pagination"},
	} {
		for _, offset := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s/offset-%d", tc.name, offset), func(t *testing.T) {
				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					if offset == 1 && requests == 1 {
						_, _ = fmt.Fprint(w, tokenPageXML(2, 0, 1))
						return
					}
					assert.Equal(t, strconv.Itoa(offset), r.URL.Query().Get("offset"))
					body := strings.ReplaceAll(tokenPageXML(2, offset, 1), "<opensearch:totalResults>2</opensearch:totalResults>", tc.total)
					_, _ = fmt.Fprint(w, body)
				}))
				defer server.Close()

				e, err := newTestAuth(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "foo")
				require.ErrorContains(t, err, tc.want)
				require.ErrorContains(t, err, fmt.Sprintf("offset %d", offset))
				require.Nil(t, e)
				require.Equal(t, offset+1, requests)
			})
		}
	}
}

func TestAuthPaginationCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "0" {
			_, _ = fmt.Fprint(w, tokenPageXML(2, 0, 1))
			return
		}
		cancel()
		<-r.Context().Done()
	}))
	defer server.Close()
	e, err := newTestAuth(ctx, componenttest.NewNopTelemetrySettings(), server.URL, "foo")
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, e)
}

func tokenPageXML(total, offset, count int) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, `<feed xmlns="http://www.w3.org/2005/Atom" xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/" xmlns:s="http://dev.splunk.com/ns/rest"><opensearch:totalResults>%d</opensearch:totalResults><opensearch:itemsPerPage>%d</opensearch:itemsPerPage><opensearch:startIndex>%d</opensearch:startIndex>`, total, count, offset)
	for i := offset; i < offset+count; i++ {
		_, _ = fmt.Fprintf(&b, `<entry><content><s:dict><s:key name="token">token-%d</s:key><s:key name="index">index-%d</s:key><s:key name="indexes"><s:list><s:item>index-%d</s:item><s:item>main</s:item></s:list></s:key></s:dict></content></entry>`, i, i, i)
	}
	b.WriteString("</feed>")
	return b.String()
}

func TestTokenSourceReloadsAllPages(t *testing.T) {
	var generation atomic.Int32
	offsets := make(chan string, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsets <- r.URL.Query().Get("offset")
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			http.Error(w, "invalid offset", http.StatusBadRequest)
			return
		}
		body := strings.ReplaceAll(tokenPageXML(2, offset, 1), "token-", fmt.Sprintf("generation-%d-token-", generation.Load()))
		_, _ = fmt.Fprint(w, body)
	}))
	defer server.Close()
	source, err := newTokenSource(t.Context(), &Config{Endpoint: server.URL, SessionKey: "session-key"})
	require.NoError(t, err)
	for i := range int32(2) {
		generation.Store(i)
		tokens, err := source.Load(t.Context())
		require.NoError(t, err)
		require.Len(t, tokens, 2)
		for j, token := range tokens {
			require.Equal(t, fmt.Sprintf("generation-%d-token-%d", i, j), string(token.Token))
		}
		require.Equal(t, "0", <-offsets)
		require.Equal(t, "1", <-offsets)
	}
}
