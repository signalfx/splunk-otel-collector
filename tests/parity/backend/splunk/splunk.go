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

// Package splunk is a parity Backend backed by a real Splunk Enterprise running
// in Docker via testcontainers. Both agents forward into it (the collector over
// HEC /event, UF over httpout /s2s), and validation is done by querying indexed
// events over the REST search API. This mirrors the boot recipe and search
// client used by splunk-otel-collector-chart's test suite.
package splunk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/signalfx/splunk-otel-collector/tests/parity"
)

const (
	defaultImage    = "splunk/splunk:9.4.4"
	defaultPassword = "helloworld"
	// readyLog is the entrypoint's completion marker.
	readyLog = "Ansible playbook complete"
)

// Config configures the Splunk backend.
type Config struct {
	Image          string        // default splunk/splunk:9.4.4
	Password       string        // admin password, default "helloworld"
	Token          string        // HEC token; must be >= 36 chars for UF httpout
	Indexes        []string      // indexes to create at startup
	StartupTimeout time.Duration // default 5m (first pull + boot)
}

func (c Config) withDefaults() Config {
	if c.Image == "" {
		c.Image = defaultImage
	}
	if c.Password == "" {
		c.Password = defaultPassword
	}
	if c.Token == "" {
		// A valid, non-zero GUID: UF httpout rejects the all-zeros GUID.
		c.Token = "11111111-1111-1111-1111-111111111111"
	}
	if c.StartupTimeout == 0 {
		c.StartupTimeout = 5 * time.Minute
	}
	return c
}

// Splunk is a Dockerized Splunk parity Backend.
type Splunk struct {
	container testcontainers.Container
	client    *http.Client
	host      string
	hecPort   string
	mgmtPort  string
	cfg       Config
}

var _ parity.Backend = (*Splunk)(nil)

// New returns an unstarted Splunk backend.
func New(cfg Config) *Splunk {
	return &Splunk{
		cfg: cfg.withDefaults(),
		client: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec // G402: the Splunk container uses a self-signed cert
		},
	}
}

func (s *Splunk) Start(ctx context.Context) error {
	req := testcontainers.ContainerRequest{
		Image: s.cfg.Image,
		// Splunk publishes amd64 only; on arm64 hosts Docker runs it under
		// emulation. Pinning the platform avoids a "no matching manifest" pull
		// error on Apple Silicon.
		ImagePlatform: "linux/amd64",
		ExposedPorts:  []string{"8088/tcp", "8089/tcp"},
		Env: map[string]string{
			"SPLUNK_START_ARGS": "--accept-license",
			"SPLUNK_PASSWORD":   s.cfg.Password,
			"SPLUNK_HEC_TOKEN":  s.cfg.Token,
		},
		WaitingFor: wait.ForLog(readyLog).WithStartupTimeout(s.cfg.StartupTimeout),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("start splunk container: %w", err)
	}
	s.container = c

	if s.host, err = c.Host(ctx); err != nil {
		return err
	}
	hp, err := c.MappedPort(ctx, "8088/tcp")
	if err != nil {
		return err
	}
	s.hecPort = hp.Port()
	mp, err := c.MappedPort(ctx, "8089/tcp")
	if err != nil {
		return err
	}
	s.mgmtPort = mp.Port()

	for _, idx := range s.cfg.Indexes {
		if err := s.createIndex(ctx, idx); err != nil {
			return fmt.Errorf("create index %s: %w", idx, err)
		}
	}
	return nil
}

func (s *Splunk) Stop(ctx context.Context) error {
	if s.container == nil {
		return nil
	}
	return s.container.Terminate(ctx)
}

func (s *Splunk) HEC() parity.HEC {
	return parity.HEC{
		Endpoint: fmt.Sprintf("https://%s:%s", s.host, s.hecPort),
		Token:    s.cfg.Token,
	}
}

func (s *Splunk) mgmtURL() string {
	return fmt.Sprintf("https://%s:%s", s.host, s.mgmtPort)
}

func (s *Splunk) createIndex(ctx context.Context, name string) error {
	form := url.Values{"name": {name}, "output_mode": {"json"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mgmtURL()+"/services/data/indexes", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", s.cfg.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	// 201 created; 409 already exists. Both are fine.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}
	return nil
}

// Search runs spl over all time and returns the matching events as Records.
func (s *Splunk) Search(ctx context.Context, spl string) ([]parity.Record, error) {
	sid, err := s.postSearch(ctx, spl)
	if err != nil {
		return nil, err
	}
	if err := s.waitJobDone(ctx, sid); err != nil {
		return nil, err
	}
	return s.jobEvents(ctx, sid)
}

// Clean deletes the events matched by spl. Best-effort: requires the admin role
// to hold the "delete" capability. With per-agent indexes it is usually
// unnecessary.
func (s *Splunk) Clean(ctx context.Context, spl string) error {
	_, err := s.Search(ctx, spl+" | delete")
	return err
}

func (s *Splunk) postSearch(ctx context.Context, spl string) (string, error) {
	form := url.Values{
		"search":        {spl},
		"earliest_time": {"0"},
		"latest_time":   {"now"},
		"output_mode":   {"json"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.mgmtURL()+"/services/search/jobs", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("admin", s.cfg.Password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var jr struct {
		Sid string `json:"sid"`
	}
	if err := json.Unmarshal(body, &jr); err != nil {
		return "", fmt.Errorf("parse job response (%d): %s", resp.StatusCode, body)
	}
	if jr.Sid == "" {
		return "", fmt.Errorf("no sid (%d): %s", resp.StatusCode, body)
	}
	return jr.Sid, nil
}

func (s *Splunk) waitJobDone(ctx context.Context, sid string) error {
	statusURL := s.mgmtURL() + "/services/search/jobs/" + sid + "?output_mode=json"
	for i := 0; i < 30; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, http.NoBody)
		if err != nil {
			return err
		}
		req.SetBasicAuth("admin", s.cfg.Password)
		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var sr struct {
			Entry []struct {
				Content struct {
					IsDone bool `json:"isDone"`
				} `json:"content"`
			} `json:"entry"`
		}
		if err := json.Unmarshal(body, &sr); err == nil && len(sr.Entry) > 0 && sr.Entry[0].Content.IsDone {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("search job %s did not finish", sid)
}

func (s *Splunk) jobEvents(ctx context.Context, sid string) ([]parity.Record, error) {
	eventsURL := s.mgmtURL() + "/services/search/jobs/" + sid + "/events?output_mode=json&count=0"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, eventsURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("admin", s.cfg.Password)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var er struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(body, &er); err != nil {
		return nil, fmt.Errorf("parse events: %w", err)
	}
	records := make([]parity.Record, 0, len(er.Results))
	for _, r := range er.Results {
		records = append(records, resultToRecord(r))
	}
	return records, nil
}

// resultToRecord maps a Splunk search result row to a parity.Record. Splunk
// returns field values as strings (or []string for multivalued). Internal
// volatile fields (_bkt, _cd, _indextime, splunk_server, ...) are left out of
// Record, so the validator never sees them.
func resultToRecord(r map[string]any) parity.Record {
	rec := parity.Record{
		Raw:        str(r["_raw"]),
		Host:       str(r["host"]),
		Source:     str(r["source"]),
		Sourcetype: str(r["sourcetype"]),
		Index:      str(r["index"]),
	}
	if t := str(r["_time"]); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			rec.Time = parsed
		}
	}
	return rec
}

func str(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		if len(x) > 0 {
			return str(x[0])
		}
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}
