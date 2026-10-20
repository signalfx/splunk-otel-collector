// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Seeds and exercises more than two default HEC token pages (OTL-4676).
//
// Invoked by test-ip-sans.sh as `go run . seed INPUTS_CONF` or
// `go run . check CA_CERT` with TEST_HEC_PAGINATION=1. All tokens are
// configured before Splunk starts because the TA reads tokens only once, at
// startup.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const tokenCount = 61

// composeArgs are the docker compose arguments (excluding the docker binary
// itself) used to run commands inside the splunk container.
var composeArgs = []string{
	"compose",
	"-f", "docker-compose.yml",
	"-f", "docker-compose.test.yml",
	"--profile", "splunk",
}

func tokenName(i int) string {
	return fmt.Sprintf("http://pagination-%03d", i)
}

// uuidFromInt renders n as a canonical UUID string, matching Python's
// uuid.UUID(int=n).
func uuidFromInt(n int) string {
	hex := fmt.Sprintf("%032x", n)
	return hex[:8] + "-" + hex[8:12] + "-" + hex[12:16] + "-" + hex[16:20] + "-" + hex[20:]
}

func seed(inputsConf string) error {
	f, err := os.OpenFile(inputsConf, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n[http]\ndisabled = 0\n"); err != nil {
		return err
	}
	for i := range tokenCount {
		stanza := fmt.Sprintf(
			"\n[%s]\ntoken = %s\nindex = main\nindexes = main\ndisabled = 0\n",
			tokenName(i), uuidFromInt(i+1),
		)
		if _, err := f.WriteString(stanza); err != nil {
			return err
		}
	}
	return nil
}

type hecEntry struct {
	Name    string `json:"name"`
	Content struct {
		Token string `json:"token"`
	} `json:"content"`
}

// readTokenPage lists HEC tokens via the same management API endpoint as
// internal/auth.NewTokenSource, using the given query string to control pagination.
func readTokenPage(query string) ([]hecEntry, error) {
	shellCmd := fmt.Sprintf(
		`curl -fsSk --max-time 30 -u "admin:$SPLUNK_PASSWORD" "https://localhost:8089/servicesNS/-/-/data/inputs/http?%s"`,
		query,
	)
	args := append(append([]string{}, composeArgs...), "exec", "-T", "splunk", "sh", "-c", shellCmd)
	out, err := exec.Command("docker", args...).Output() //nolint:gosec // test fixture, fixed args plus a literal query string
	if err != nil {
		return nil, fmt.Errorf("list HEC tokens (%s): %w", query, err)
	}
	var page struct {
		Entry []hecEntry `json:"entry"`
	}
	if err := json.Unmarshal(out, &page); err != nil {
		return nil, fmt.Errorf("parse HEC token listing (%s): %w", query, err)
	}
	return page.Entry, nil
}

func statusFor(token, caCert string) (string, error) {
	out, err := exec.Command( //nolint:gosec // test fixture, args are not externally controlled
		"curl", "-sS", "--max-time", "10", "--cacert", caCert,
		"-o", "/dev/null", "-w", "%{http_code}",
		"-H", "Authorization: Splunk "+token,
		"-H", "Content-Type: application/json", "-d", "{}",
		"https://localhost:4318/v1/logs",
	).Output()
	return string(out), err
}

func check(caCert string) error {
	tokenNames := make(map[string]bool, tokenCount)
	for i := range tokenCount {
		tokenNames[tokenName(i)] = true
	}

	// Use the same default listing as internal/auth.NewTokenSource, plus an unlimited
	// listing to prove that valid tokens really were omitted from that page.
	firstPage, err := readTokenPage("output_mode=json")
	if err != nil {
		return err
	}
	allTokens, err := readTokenPage("output_mode=json&count=0")
	if err != nil {
		return err
	}

	tokens := make(map[string]string)
	for _, e := range allTokens {
		if tokenNames[e.Name] {
			tokens[e.Name] = e.Content.Token
		}
	}
	firstNames := make(map[string]bool)
	for _, e := range firstPage {
		if tokenNames[e.Name] {
			firstNames[e.Name] = true
		}
	}
	omitted := 0
	for name := range tokenNames {
		if !firstNames[name] {
			omitted++
		}
	}
	if len(tokens) != tokenCount || len(firstNames) == 0 || omitted < 31 {
		return fmt.Errorf(
			"pagination fixture invalid: %d seeded tokens visible, %d on the default page, %d omitted",
			len(tokens), len(firstNames), omitted,
		)
	}

	// A negative control ensures the receiver actually enforces authentication.
	status, err := statusFor(uuidFromInt(0), caCert)
	if err != nil {
		return err
	}
	if status != "401" {
		return errors.New("OTLP receiver did not reject an unknown HEC token")
	}

	names := make([]string, 0, len(tokens))
	for name := range tokens {
		names = append(names, name)
	}
	sort.Strings(names)

	var failures []string
	for _, name := range names {
		status, err := statusFor(tokens[name], caCert)
		if err != nil {
			return err
		}
		if status != "200" {
			page := "later page"
			if firstNames[name] {
				page = "default page"
			}
			failures = append(failures, fmt.Sprintf("%s (%s): expected HTTP 200, got %s", name, page, status))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("HEC token pagination regression:\n%s", strings.Join(failures, "\n"))
	}

	fmt.Printf(
		"OK: all %d HEC tokens authenticated, including %d omitted from the default page\n",
		tokenCount, omitted,
	)
	return nil
}

func main() {
	if len(os.Args) != 3 || (os.Args[1] != "seed" && os.Args[1] != "check") {
		fmt.Fprintf(os.Stderr, "Usage: %s seed INPUTS_CONF | check CA_CERT\n", os.Args[0])
		os.Exit(1)
	}

	var err error
	if os.Args[1] == "seed" {
		err = seed(os.Args[2])
	} else {
		err = check(os.Args[2])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
