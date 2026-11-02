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

package health

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component/componentstatus"

	"github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"
)

const verbName = "health"

const (
	defaultHost = "localhost"
	defaultPort = 13133
	defaultPath = "/status"

	requestTimeout = 5 * time.Second
)

const usageText = `Query the healthcheckv2 extension's HTTP status endpoint and print the
collector's overall health along with the health of each reporting component.

By default this queries http://localhost:13133/status?verbose, matching the
healthcheckv2 extension's own default HTTP host, port, and status path.

Flags:
  --host string        Host the healthcheckv2 extension's HTTP status server is
                       listening on (default "localhost").
  --port int           Port the healthcheckv2 extension's HTTP status server is
                       listening on (default 13133).
  --path string        Path of the status endpoint, matching the extension's
                       http.status.path setting (default "/status").
  --pipeline string    Restrict the query to a single pipeline (e.g. "traces"
                       or "metrics/grpc") instead of overall collector health.
  --component string   Only print components whose name contains this
                       substring (e.g. "otlp" or "receiver:otlp").

Examples:
  otelcollauncher health
  otelcollauncher health --port 13140
  otelcollauncher health --pipeline traces/http
  otelcollauncher health --component exporter
`

type Family struct{}

// New returns the health command family.
func New() *Family {
	return &Family{}
}

func (Family) Verbs() []cli.VerbInfo {
	return []cli.VerbInfo{
		{
			Name:  verbName,
			Help:  "query the healthcheckv2 extension and print collector health",
			Usage: usageText,
		},
	}
}

func (Family) Dispatch(ctx context.Context, verb string, args []string, streams cli.IO) int {
	if verb != verbName {
		fmt.Fprintf(streams.Err, "unknown health command %q\n", verb)
		return cli.ExitUsage
	}
	return run(ctx, args, streams)
}

type options struct {
	host      string
	path      string
	pipeline  string
	component string
	port      int
}

func parseArgs(args []string, streams cli.IO) (options, error) {
	fs := flag.NewFlagSet(verbName, flag.ContinueOnError)
	fs.SetOutput(streams.Err)
	fs.Usage = func() { fmt.Fprint(streams.Err, usageText) }

	var opts options
	fs.StringVar(&opts.host, "host", defaultHost, "healthcheckv2 HTTP status host")
	fs.IntVar(&opts.port, "port", defaultPort, "healthcheckv2 HTTP status port")
	fs.StringVar(&opts.path, "path", defaultPath, "healthcheckv2 status endpoint path")
	fs.StringVar(&opts.pipeline, "pipeline", "", "restrict the query to a single pipeline")
	fs.StringVar(&opts.component, "component", "", "filter output to components matching this substring")

	if err := fs.Parse(args); err != nil {
		return options{}, err
	}
	if fs.NArg() > 0 {
		err := fmt.Errorf("health: unexpected argument %q", fs.Arg(0))
		fmt.Fprintf(streams.Err, "%v\n\n", err)
		fmt.Fprint(streams.Err, usageText)
		return options{}, err
	}
	return opts, nil
}

func run(ctx context.Context, args []string, streams cli.IO) int {
	opts, err := parseArgs(args, streams)
	if err != nil {
		return cli.ExitUsage
	}

	target := buildURL(opts)
	status, err := fetchStatus(ctx, target)
	if err != nil {
		if ctx.Err() != nil {
			fmt.Fprintf(streams.Err, "health query interrupted: %v\n", ctx.Err())
			return cli.ExitFailure
		}
		// a non-200 response can return both a parsed status and an error
		if status != nil {
			render(streams.Out, target, status, opts.component)
			fmt.Fprintln(streams.Err, err)
			return cli.ExitFailure
		}
		fmt.Fprintf(streams.Err, "%v\n\n%s", err, connectionHint(opts))
		return connectionExitCode(err)
	}
	if status == nil {
		fmt.Fprintln(streams.Err, "health query returned no status")
		return cli.ExitFailure
	}

	render(streams.Out, target, status, opts.component)
	if !status.Healthy {
		return cli.ExitFailure
	}
	return cli.ExitOK
}

func buildURL(opts options) string {
	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(opts.host, strconv.Itoa(opts.port)),
		Path:   opts.path,
	}
	query := "verbose"
	if opts.pipeline != "" {
		query = "pipeline=" + url.QueryEscape(opts.pipeline) + "&" + query
	}
	u.RawQuery = query
	return u.String()
}

func fetchStatus(ctx context.Context, target string) (*componentStatus, error) {
	client := &http.Client{Timeout: requestTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach %s: %w", target, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from %s: %w", target, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s returned 404 Not Found (an unknown --pipeline name returns 404)", target)
	}

	var status componentStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("%s did not return a valid healthcheckv2 status response: %w", target, err)
	}
	if err := status.validate(); err != nil {
		return nil, fmt.Errorf("%s did not return a valid healthcheckv2 status response: %w", target, err)
	}

	// non-200 response can still return content status; see extension code for details internal/httpserver/responders.go
	if resp.StatusCode != http.StatusOK {
		return &status, fmt.Errorf("%s returned HTTP %s", target, resp.Status)
	}
	return &status, nil
}

func connectionHint(opts options) string {
	return fmt.Sprintf(
		"Make sure the healthcheckv2 extension is enabled and its HTTP status\n"+
			"service is reachable at %s:%d%s. If it's configured with a different\n"+
			"host, port, or path, pass --host, --port, and/or --path to match.\n",
		opts.host, opts.port, opts.path,
	)
}

func connectionExitCode(err error) int {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return cli.ExitFailure
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" && !opErr.Timeout() {
		return cli.ExitNotRunning
	}
	return cli.ExitFailure
}

type componentStatus struct {
	Timestamp      time.Time                   `json:"status_time"`
	StartTimestamp *time.Time                  `json:"start_time,omitempty"`
	Attributes     map[string]any              `json:"attributes,omitempty"`
	Components     map[string]*componentStatus `json:"components,omitempty"`
	Status         string                      `json:"status"`
	Error          string                      `json:"error,omitempty"`
	Healthy        bool                        `json:"healthy"`
}

var validStatuses = func() map[string]bool {
	m := make(map[string]bool, componentstatus.StatusStopped-componentstatus.StatusNone+1)
	for s := componentstatus.StatusNone; s <= componentstatus.StatusStopped; s++ {
		m[s.String()] = true
	}
	return m
}()

func (s *componentStatus) validate() error {
	if !validStatuses[s.Status] {
		return fmt.Errorf("missing or unrecognized \"status\" field (got %q)", s.Status)
	}
	if s.Timestamp.IsZero() {
		return errors.New(`missing "status_time" field`)
	}
	for name, cs := range s.Components {
		if cs == nil {
			return fmt.Errorf("component %q is null", name)
		}
		if err := cs.validate(); err != nil {
			return fmt.Errorf("component %q: %w", name, err)
		}
	}
	return nil
}

func (s *componentStatus) shortStatus() string {
	return strings.TrimPrefix(s.Status, "Status")
}

func render(w io.Writer, target string, status *componentStatus, filter string) {
	fmt.Fprintf(w, "Queried: %s\n", target)
	fmt.Fprintf(w, "Overall status: %s (healthy: %s)\n", status.shortStatus(), healthyLabel(status.Healthy))
	if status.Error != "" {
		fmt.Fprintf(w, "Error: %s\n", status.Error)
	}
	if status.StartTimestamp != nil {
		fmt.Fprintf(w, "Started: %s\n", status.StartTimestamp.Format(time.RFC3339))
	}
	fmt.Fprintf(w, "As of: %s\n", status.Timestamp.Format(time.RFC3339))

	if len(status.Components) == 0 {
		return
	}

	rows := collectRows(status.Components, filter, 0)
	fmt.Fprintln(w)
	if len(rows) == 0 {
		fmt.Fprintf(w, "No components matched --component %q\n", filter)
		return
	}

	printTable(w, rows)
}

type componentRow struct {
	label   string
	status  string
	healthy string
	err     string
	depth   int
}

func collectRows(components map[string]*componentStatus, filter string, depth int) []componentRow {
	var rows []componentRow
	for _, name := range filterComponents(components, filter) {
		cs := components[name]
		if cs == nil {
			continue
		}
		rows = append(rows, componentRow{
			depth:   depth,
			label:   name,
			status:  cs.shortStatus(),
			healthy: healthyLabel(cs.Healthy),
			err:     cs.Error,
		})
		rows = append(rows, collectRows(cs.Components, filter, depth+1)...)
	}
	return rows
}

func filterComponents(components map[string]*componentStatus, filter string) []string {
	names := make([]string, 0, len(components))
	for name, cs := range components {
		if componentMatches(name, cs, filter) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func componentMatches(name string, cs *componentStatus, filter string) bool {
	if filter == "" || strings.Contains(strings.ToLower(name), strings.ToLower(filter)) {
		return true
	}
	if cs == nil {
		return false
	}
	for childName, child := range cs.Components {
		if componentMatches(childName, child, filter) {
			return true
		}
	}
	return false
}

func printTable(w io.Writer, rows []componentRow) {
	const headerLabel, headerStatus, headerHealthy = "COMPONENT", "STATUS", "HEALTHY"

	labelWidth, statusWidth := len(headerLabel), len(headerStatus)
	for _, r := range rows {
		labelWidth = max(labelWidth, r.depth*2+len(r.label))
		statusWidth = max(statusWidth, len(r.status))
	}

	fmt.Fprintf(w, "%-*s  %-*s  %s\n", labelWidth, headerLabel, statusWidth, headerStatus, headerHealthy)
	for _, r := range rows {
		indent := strings.Repeat("  ", r.depth)
		fmt.Fprintf(w, "%-*s  %-*s  %s\n", labelWidth, indent+r.label, statusWidth, r.status, r.healthy)
		if r.err != "" {
			fmt.Fprintf(w, "%s  error: %s\n", indent, r.err)
		}
	}
}

func healthyLabel(healthy bool) string {
	if healthy {
		return "yes"
	}
	return "no"
}
