// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/signalfx/splunk-otel-collector/internal/auth/authtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 5 * time.Second
)

func TestMainPrintsScheme(t *testing.T) {
	output := CaptureStdout(t, func() {
		originalArgs := os.Args
		os.Args = []string{"splunk-connect-for-otlp", "--scheme"}
		defer func() {
			os.Args = originalArgs
		}()

		main()
	})

	require.Equal(t, Scheme+"\n", output)
}

func TestRunReturnsErrorForInvalidInput(t *testing.T) {
	restoreStdin := WriteToStdin(t, "not-xml")
	defer restoreStdin()

	err := run()
	require.Error(t, err)
}

func TestRunStartsAndStopsOnSignal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	authtest.SetupAuth(listener, t)

	restoreStdin := WriteToStdin(t, fmt.Sprintf(
		`<input>
  <server_uri>http://%s</server_uri>
  <session_key>mysessionkey</session_key>
<configuration>
  <stanza name="splunk-connect-for-otlp://test" app="search">
    <param name="grpc_port">0</param>
    <param name="http_port">0</param>
    <param name="listen_address">127.0.0.1</param>
    <param name="enableSSL">0</param>
  </stanza>
</configuration>
</input>`, listener.Addr().String(),
	))
	defer restoreStdin()

	done := make(chan error, 1)
	go func() {
		done <- run()
	}()

	time.AfterFunc(500*time.Millisecond, func() {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("run did not complete in time")
	}
}

func TestExpectedHEC(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	authtest.SetupAuth(listener, t)
	serverCert := filepath.Join("testdata", "cert.pem")
	serverKey := filepath.Join("testdata", "key.pem")

	tests := []struct {
		name         string
		otlpendpoint string
		inputPath    string
		expectedPath string
	}{
		{
			name:         "metrics",
			otlpendpoint: "/v1/metrics",
			inputPath:    filepath.Join("testdata", "otlp_metrics.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_metrics.json"),
		},
		{
			name:         "traces",
			otlpendpoint: "/v1/traces",
			inputPath:    filepath.Join("testdata", "otlp_traces.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_traces.json"),
		},
		{
			name:         "logs",
			otlpendpoint: "/v1/logs",
			inputPath:    filepath.Join("testdata", "otlp_logs.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_logs.json"),
		},
		{
			name:         "large file metrics",
			otlpendpoint: "/v1/metrics",
			inputPath:    filepath.Join("testdata", "otlp_metrics_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_metrics_big.json"),
		},
		{
			name:         "large file traces",
			otlpendpoint: "/v1/traces",
			inputPath:    filepath.Join("testdata", "otlp_traces_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_traces_big.json"),
		},
		{
			name:         "large file logs",
			otlpendpoint: "/v1/logs",
			inputPath:    filepath.Join("testdata", "otlp_logs_big.json"),
			expectedPath: filepath.Join("testdata", "expected_hec_logs_big.json"),
		},
	}

	for _, tt := range tests {
		tt := tt
		for _, ssl := range []bool{true, false} {
			name := tt.name
			if ssl {
				name += "-ssl"
			}
			t.Run(name, func(t *testing.T) {
				grpcPort := GetFreePort(t)
				httpPort := GetFreePort(t)
				config := fmt.Sprintf(`<input>
  <server_uri>http://%s</server_uri>
  <session_key>mysessionkey</session_key>
<configuration><stanza name="splunk-connect-for-otlp://test" app="search">
<param name="grpc_port">%d</param>
<param name="http_port">%d</param>
<param name="enableSSL">%v</param>
<param name="serverCert">%s</param>
<param name="serverKey">%s</param>
<param name="listen_address">127.0.0.1</param>
</stanza></configuration></input>`, listener.Addr().String(), grpcPort, httpPort, ssl, serverCert, serverKey)

				restoreStdin := WriteToStdin(t, config)
				t.Cleanup(restoreStdin)

				stdoutLines, restoreStdout := CaptureStdoutLines(t)
				t.Cleanup(restoreStdout)

				runDone := make(chan error, 1)
				go func() {
					err := run()
					require.NoError(t, err)
					runDone <- err
				}()
				// wait until the receiver is up.
				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", httpPort), 100*time.Millisecond)
					require.NoError(tt, err)
					_ = conn.Close()
				}, 5*time.Second, 200*time.Millisecond)

				payload, err := os.ReadFile(tt.inputPath)
				require.NoError(t, err)

				expected := LoadExpectedHecData(t, tt.expectedPath)
				expectedLines := strings.Split(strings.TrimSpace(string(expected)), "\n")
				require.NotEmpty(t, expectedLines, "%s must contain fixture data", tt.expectedPath)

				PostOTLP(t, httpPort, tt.otlpendpoint, payload, ssl)

				actual := CollectLines(t, stdoutLines, len(expectedLines))
				require.Equal(t, expectedLines, actual)

				require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))
				require.NoError(t, <-runDone)
			})
		}
	}
}

func GetFreePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on ephemeral port: %v", err)
	}
	defer func() {
		err = l.Close()
		if err != nil {
			t.Fatalf("failed to close ephemeral port: %v", err)
		}
	}()
	return l.Addr().(*net.TCPAddr).Port
}

// PostOTLP sends the provided data to the endpoint, retrying until success or timeout.
func PostOTLP(t *testing.T, port int, path string, body []byte, ssl bool) {
	t.Helper()

	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
	if ssl {
		url = fmt.Sprintf("https://127.0.0.1:%d%s", port, path)
	}
	deadline := time.Now().Add(defaultTimeout)

	lastRespCode := 0
	for {
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Splunk 00000000-0000-0000-0000-0000000000000")
		var client *http.Client
		if ssl {
			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
		} else {
			client = http.DefaultClient
		}
		resp, err := client.Do(req)
		if err == nil {
			lastRespCode = resp.StatusCode
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		} else {
			require.NoError(t, err)
		}

		if time.Now().After(deadline) {
			t.Fatalf("failed to POST %s, response code: %v", path, lastRespCode)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func WriteToStdin(t *testing.T, content string) func() {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	if _, err = io.Copy(w, strings.NewReader(content)); err != nil {
		t.Fatalf("failed to write stdin content: %v", err)
	}
	if err = w.Close(); err != nil {
		t.Fatalf("failed to close stdin writer: %v", err)
	}

	original := os.Stdin
	os.Stdin = r

	return func() {
		os.Stdin = original
		_ = r.Close()
	}
}

// CollectLines reads expectedCount lines from the provided channel or fails on timeout.
// Note: There's the possibility that more lines were sent to stdout than expected,
// so the caller must check returned data contents.
func CollectLines(t *testing.T, ch <-chan string, expectedCount int) []string {
	t.Helper()

	var lines []string
	timeout := time.After(defaultTimeout)
	for len(lines) < expectedCount {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatalf("stdout closed early, got %d lines, expected %d", len(lines), expectedCount)
			}
			lines = append(lines, line)
		case <-timeout:
			t.Fatalf("timed out waiting for stdout lines; got %d expected %d", len(lines), expectedCount)
		}
	}
	return lines
}

func LoadExpectedHecData(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read expected data %s: %v", path, err)
	}
	return data
}

// CaptureStdout captures all stdout output generated by fn and returns it as a string.
func CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outputCh <- buf.String()
	}()

	fn()

	// TODO: Find a way to synchronize without sleep.
	time.Sleep(1 * time.Second)

	_ = w.Close()
	os.Stdout = original

	return <-outputCh
}

// CaptureStdoutLines returns a channel streaming stdout lines and a restore function.
func CaptureStdoutLines(t *testing.T) (<-chan string, func()) {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w

	lines := make(chan string, 50)
	go func() {
		scanner := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
		_ = r.Close()
	}()

	return lines, func() {
		os.Stdout = original
		_ = w.Close()
	}
}
