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
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensionauth"

	"github.com/signalfx/splunk-otel-collector/internal/auth/authtest"
)

func TestAuth(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	authtest.SetupAuth(t, listener)
	e, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), "http://"+listener.Addr().String(), "foo")
	require.NoError(t, err)
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

func TestAuthSucceedsForLoopbackTLSCertificateWithoutIPSANs(t *testing.T) {
	server := authtest.NewTLSServer(t)

	_, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "foo")
	require.NoError(t, err)
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
				require.Equal(t, "/servicesNS/-/-/data/inputs/http", r.URL.Path)
				require.Equal(t, "Splunk session-key", r.Header.Get("Authorization"))
				require.False(t, r.URL.Query().Has("count"))
				offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
				require.NoError(t, err)
				offsets = append(offsets, offset)
				_, _ = fmt.Fprint(w, tokenPageXML(tc.total, offset, min(tc.pageSize, tc.total-offset)))
			}))
			defer server.Close()

			e, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "session-key")
			require.NoError(t, err)
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
				require.Equal(t, "1", r.URL.Query().Get("offset"))
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			e, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "foo")
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
					require.Equal(t, strconv.Itoa(offset), r.URL.Query().Get("offset"))
					body := strings.ReplaceAll(tokenPageXML(2, offset, 1), "<opensearch:totalResults>2</opensearch:totalResults>", tc.total)
					_, _ = fmt.Fprint(w, body)
				}))
				defer server.Close()

				e, err := New(context.Background(), componenttest.NewNopTelemetrySettings(), server.URL, "foo")
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
	e, err := New(ctx, componenttest.NewNopTelemetrySettings(), server.URL, "foo")
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
