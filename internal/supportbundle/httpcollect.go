// Copyright Splunk Inc.
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

// This file collects everything scraped from the collector's own HTTP
// endpoints: the prometheus self-metrics page and the zpages extension.

package supportbundle

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

const httpFetchTimeout = 5 * time.Second

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
	},
}

func fetchBody(ctx context.Context, rawURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, httpFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func fetchToFile(ctx context.Context, rawURL, dest string) error {
	body, err := fetchBody(ctx, rawURL)
	if err != nil {
		return err
	}
	return writeBodyToFile(ctx, dest, body)
}

func writeBodyToFile(ctx context.Context, dest string, body []byte) (err error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	f, err := createOutputFile(dest)
	if err != nil {
		return err
	}
	defer func() {
		err = finishOutputFile(f, dest, err)
	}()

	if _, err = f.Write(body); err != nil {
		return err
	}
	return ctx.Err()
}

const metricsURL = "http://localhost:8888/metrics"

func collectMetrics(ctx context.Context, logger *logger, tempDir string) {
	logger.Info("Getting metric information...")
	dest := filepath.Join(tempDir, metricsDirName, "collector-metrics.txt")
	if err := fetchToFile(ctx, metricsURL, dest); err != nil {
		logger.Warnf("localhost:8888/metrics unavailable so metrics not collected: %v", err)
	}
}

const expvarzURL = "http://localhost:55679/debug/expvarz"

func collectZpages(ctx context.Context, logger *logger, tempDir string) {
	logger.Info("Getting expvarz information...")
	zpagesDir := filepath.Join(tempDir, zpagesDirName)

	if err := fetchToFile(ctx, expvarzURL, filepath.Join(zpagesDir, "expvarz.json")); err != nil {
		logger.Warnf("failed to collect expvarz: %v", err)
	}
}
