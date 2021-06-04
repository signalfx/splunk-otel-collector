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
	"log"
	"net/url"
	"strings"
)

type saCfgInfo struct {
	accessToken string
	realm       string
	monitors    []interface{}
}

func saExpandedToCfgInfo(saExpanded map[interface{}]interface{}) saCfgInfo {
	return saCfgInfo{
		accessToken: saExpanded["signalFxAccessToken"].(string),
		realm:       apiURLToRealm(saExpanded["apiUrl"].(string)),
		monitors:    saExpanded["monitors"].([]interface{}),
	}
}

func apiURLToRealm(ingestURL string) string {
	u, err := url.Parse(ingestURL)
	if err != nil {
		log.Fatalf("failed to get realm: %v", err)
	}

	host := strings.ToLower(u.Host)
	if host == "api.signalfx.com" {
		return "us0"
	}

	parts := strings.Split(u.Host, ".")
	realm := parts[1]
	return realm
}
