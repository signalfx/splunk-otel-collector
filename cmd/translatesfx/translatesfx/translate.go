// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type saCfgInfo struct {
	accessToken string
	realm       string
	monitors    []interface{}
}

func saExpandedToCfgInfo(saExpanded map[interface{}]interface{}, wd string) (saCfgInfo, error) {
	realm, err := apiURLToRealm(saExpanded["apiUrl"].(string))
	if err != nil {
		return saCfgInfo{}, err
	}
	return saCfgInfo{
		accessToken: saExpanded["signalFxAccessToken"].(string),
		realm:       realm,
		monitors:    saExpanded["monitors"].([]interface{}),
	}, nil
}

func resolvePath(path, wd string) string {
	if path[:1] == string(os.PathSeparator) {
		return path
	}
	return filepath.Join(wd, path)
}

var includeRegexp = regexp.MustCompile(`\${include:(.*)}`)

func apiURLToRealm(ingestURL string, wd string) string {
	u := ingestURL
	if matches := includeRegexp.FindStringSubmatch(ingestURL); len(matches) == 2 {
		u = expandIncludeURL(resolvePath(matches[1], wd))
	}
	return plainURLToRealm(u)
}

func resolvePath(path, wd string) string {
	if path[:1] == string(os.PathSeparator) {
		return path
	}
	return filepath.Join(wd, path)
}

func expandIncludeURL(fname string) string {
	// TODO prepend working dir?
	bytes, err := os.ReadFile(fname)
	if err != nil {
		log.Fatalf("error reading file %s: %v", fname, err)
	}
	return string(bytes)
}

func plainURLToRealm(ingestURL string) (string, error) {
	u, err := url.Parse(ingestURL)
	if err != nil {
		return "", fmt.Errorf("failed to get realm from api %v: %v", ingestURL, err)
	}

	host := strings.ToLower(u.Host)
	if host == "api.signalfx.com" {
		return "us0", nil
	}

	parts := strings.Split(u.Host, ".")
	realm := parts[1]
	return realm, nil
}
