// Copyright Splunk, Inc.
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

package splunksecretconfigsource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSplunkSecretRetrieve(t *testing.T) {
	testsCases := []struct {
		name          string
		selector      string
		app           string
		user          string
		checkErr      func(t *testing.T, err error)
		handler       http.HandlerFunc
		expectedValue string
		expectedPath  string
		expectedAuth  string
	}{
		{
			name:         "present_with_realm",
			selector:     "myrealm:myuser",
			expectedPath: "/servicesNS/-/-/storage/passwords/myrealm:myuser",
			expectedAuth: "Splunk test_session_key",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":"s3cr3t"}}]}`))
			},
			expectedValue: "s3cr3t",
		},
		{
			name:         "present_without_realm",
			selector:     ":myuser",
			expectedPath: "/servicesNS/-/-/storage/passwords/myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":"s3cr3t2"}}]}`))
			},
			expectedValue: "s3cr3t2",
		},
		{
			name:         "present_with_specific_app",
			selector:     "myrealm:myuser",
			app:          "myapp",
			expectedPath: "/servicesNS/-/myapp/storage/passwords/myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":"s3cr3t3"}}]}`))
			},
			expectedValue: "s3cr3t3",
		},
		{
			name:         "present_with_specific_user",
			selector:     "myrealm:myuser",
			user:         "myowner",
			expectedPath: "/servicesNS/myowner/-/storage/passwords/myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":"s3cr3t4"}}]}`))
			},
			expectedValue: "s3cr3t4",
		},
		{
			name:     "ambiguous_selector_multiple_apps",
			selector: "myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":"s3cr3t"},"acl":{"app":"appone","owner":"userone"}},{"content":{"clear_password":"s3cr3t"},"acl":{"app":"apptwo","owner":"usertwo"}}]}`))
			},
			checkErr: func(t *testing.T, err error) {
				var target *errAmbiguousSelector
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "invalid_selector_no_colon",
			selector: "myuser",
			checkErr: func(t *testing.T, err error) {
				var target *errInvalidSelector
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "invalid_selector_empty_name",
			selector: "myrealm:",
			checkErr: func(t *testing.T, err error) {
				var target *errInvalidSelector
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "not_found",
			selector: "myrealm:missing",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[]}`))
			},
			checkErr: func(t *testing.T, err error) {
				var target *errNoMatchingEntry
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "missing_clear_password",
			selector: "myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"entry":[{"content":{"clear_password":""}}]}`))
			},
			checkErr: func(t *testing.T, err error) {
				var target *errMissingClearPassword
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "unexpected_status_code",
			selector: "myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"messages":[{"type":"ERROR","text":"unauthorized"}]}`))
			},
			checkErr: func(t *testing.T, err error) {
				var target *errUnexpectedStatusCode
				require.ErrorAs(t, err, &target)
			},
		},
		{
			name:     "invalid_json_response",
			selector: "myrealm:myuser",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`not json`))
			},
			checkErr: func(t *testing.T, err error) {
				var target *errRequestFailed
				require.ErrorAs(t, err, &target)
			},
		},
	}

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			var srv *httptest.Server
			if tc.handler != nil {
				srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.expectedPath != "" {
						assert.Equal(t, tc.expectedPath, r.URL.Path)
					}
					if tc.expectedAuth != "" {
						assert.Equal(t, tc.expectedAuth, r.Header.Get("Authorization"))
					}
					tc.handler(w, r)
				}))
				defer srv.Close()
			}

			cfg := &Config{SessionKey: "test_session_key"}
			if srv != nil {
				cfg.Endpoint = srv.URL
			} else {
				cfg.Endpoint = "http://127.0.0.1:0"
			}

			app := tc.app
			if app == "" {
				app = "-"
			}
			user := tc.user
			if user == "" {
				user = "-"
			}
			source := newConfigSource(cfg, app, user, time.Second, zap.NewNop())
			retrieved, err := source.Retrieve(context.Background(), tc.selector, nil, nil)
			if tc.checkErr != nil {
				require.Error(t, err)
				tc.checkErr(t, err)
				return
			}
			require.NoError(t, err)
			val, err := retrieved.AsString()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedValue, val)
		})
	}
}
