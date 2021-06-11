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
	"strings"
)

type saCfgInfo struct {
	globalDims  map[interface{}]interface{}
	accessToken string
	realm       string
	monitors    []interface{}
}

func saExpandedToCfgInfo(saExpanded map[interface{}]interface{}) (saCfgInfo, error) {
	realm, err := apiURLToRealm(saExpanded["apiUrl"].(string))
	if err != nil {
		return saCfgInfo{}, err
	}
	var globalDims map[interface{}]interface{}
	if gd, ok := saExpanded["globalDimensions"]; ok {
		globalDims = gd.(map[interface{}]interface{})
	}
	return saCfgInfo{
		accessToken: saExpanded["signalFxAccessToken"].(string),
		realm:       realm,
		monitors:    saExpanded["monitors"].([]interface{}),
		globalDims:  globalDims,
	}, nil
}

func resolvePath(path, wd string) string {
	if path[:1] == string(os.PathSeparator) {
		return path
	}
	return filepath.Join(wd, path)
}

func apiURLToRealm(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse apiURL %v: %v", apiURL, err)
	}

	host := strings.ToLower(u.Host)
	if host == "api.signalfx.com" {
		return "us0", nil
	}

	parts := strings.Split(u.Host, ".")
	realm := parts[1]
	return realm, nil
}
