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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

// Private error types to help with testability.
type (
	errInvalidSelector      struct{ error }
	errRequestFailed        struct{ error }
	errUnexpectedStatusCode struct{ error }
	errNoMatchingEntry      struct{ error }
	errAmbiguousSelector    struct{ error }
	errMissingClearPassword struct{ error }
)

// clearPasswordPatterns matches clear_password values that splunkd may echo back in an
// error response body, whether the body is JSON (output_mode=json) or the XML/Atom format
// splunkd falls back to for some error conditions (e.g. authentication failures).
var clearPasswordPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)("clear_password"\s*:\s*")[^"]*(")`),
	regexp.MustCompile(`(?i)(name="clear_password"[^>]*>)[^<]*(<)`),
}

// redactClearPassword returns body with any clear_password value replaced by a redaction
// marker, guarding against accidentally leaking a secret in an error message.
func redactClearPassword(body string) string {
	for _, re := range clearPasswordPatterns {
		body = re.ReplaceAllString(body, "${1}REDACTED${2}")
	}
	return body
}

// passwordsResponse models the subset of the storage/passwords JSON response that is
// relevant to retrieve the clear text value of a secret.
// See https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/secretstorage/secretstoragerest
type passwordsResponse struct {
	Entry []struct {
		Content struct {
			ClearPassword string `json:"clear_password"`
		} `json:"content"`
		ACL struct {
			App   string `json:"app"`
			Owner string `json:"owner"`
		} `json:"acl"`
	} `json:"entry"`
}

type splunkSecretConfigSource struct {
	client     *http.Client
	logger     *zap.Logger
	endpoint   string
	sessionKey string
	app        string
	user       string
}

func newConfigSource(cfg *Config, app, user string, timeout time.Duration, logger *zap.Logger) *splunkSecretConfigSource {
	return &splunkSecretConfigSource{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}, //nolint:gosec // opt-in, defaults to false
			},
		},
		logger:     logger,
		endpoint:   strings.TrimRight(cfg.Endpoint, "/"),
		sessionKey: cfg.SessionKey,
		app:        app,
		user:       user,
	}
}

// passwordsURL builds the storage/passwords REST endpoint URL for the given credentialID.
// See https://dev.splunk.com/enterprise/docs/developapps/manageknowledge/secretstorage/secretstoragerest and
// https://help.splunk.com/en/data-management/splunk-enterprise-admin-manual/10.4/configuration-file-reference/10.4.1-configuration-file-reference/passwords.conf
//
// servicesNS/{user}/{app}/ pins the lookup to a specific user/app namespace. A
// <realm>:<name> credential is only guaranteed unique within the user/app namespace it
// was created in, so s.user and s.app both default to "-" (all users/apps, per the
// wildcard documented at https://help.splunk.com/en/?resourceId=Splunk_RESTREF_RESTlist)
// but can be set via the User/App config fields to disambiguate credentials that collide
// across namespaces. See the "Namespace" section at
// https://help.splunk.com/en/?resourceId=Splunk_RESTUM_RESTusing for details.
func (s *splunkSecretConfigSource) passwordsURL(credentialID string) string {
	return fmt.Sprintf("%s/servicesNS/%s/%s/storage/passwords/%s?output_mode=json", s.endpoint, url.PathEscape(s.user), url.PathEscape(s.app), url.PathEscape(credentialID))
}

// Retrieve fetches the clear text value of the secret identified by selector from splunkd's
// storage/passwords REST endpoint. The selector must be in the form "[<realm>]:<name>", e.g.
// "myrealm:myuser" or ":myuser" if the secret was stored without a realm.
func (s *splunkSecretConfigSource) Retrieve(ctx context.Context, selector string, _ *confmap.Conf, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	realm, name, found := strings.Cut(selector, ":")
	if !found || name == "" {
		return nil, &errInvalidSelector{fmt.Errorf("selector %q must be in the form [<realm>]:<name>", selector)}
	}

	credentialID := name
	if realm != "" {
		credentialID = realm + ":" + name
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.passwordsURL(credentialID), http.NoBody)
	if err != nil {
		return nil, &errRequestFailed{fmt.Errorf("failed to build request for selector %q: %w", selector, err)}
	}
	req.Header.Set("Authorization", "Splunk "+s.sessionKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &errRequestFailed{fmt.Errorf("failed to retrieve secret for selector %q: %w", selector, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errRequestFailed{fmt.Errorf("failed to read response body for selector %q: %w", selector, err)}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &errUnexpectedStatusCode{fmt.Errorf("splunkd returned status %d for selector %q: %s", resp.StatusCode, selector, redactClearPassword(string(body)))}
	}

	var parsed passwordsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, &errRequestFailed{fmt.Errorf("failed to parse response for selector %q: %w", selector, err)}
	}

	if len(parsed.Entry) == 0 {
		return nil, &errNoMatchingEntry{fmt.Errorf("no secret found for selector %q", selector)}
	}
	if len(parsed.Entry) > 1 {
		owners := make([]string, 0, len(parsed.Entry))
		for _, entry := range parsed.Entry {
			owners = append(owners, entry.ACL.Owner+":"+entry.ACL.App)
		}
		return nil, &errAmbiguousSelector{fmt.Errorf("selector %q matches credentials owned by multiple user:app namespaces %v; set the User and App config fields to disambiguate", selector, owners)}
	}

	clearPassword := parsed.Entry[0].Content.ClearPassword
	if clearPassword == "" {
		return nil, &errMissingClearPassword{errors.New("secret has no clear_password value")}
	}

	return confmap.NewRetrieved(clearPassword)
}
