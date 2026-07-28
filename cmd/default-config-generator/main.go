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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

type configGenerator struct {
	basePath    string
	outputPath  string
	overlayPath string
}

func main() {
	configs := []configGenerator{
		{
			basePath:    filepath.Join("config", "agent_config.yaml"),
			overlayPath: filepath.Join("config", "agent_to_backend_overlay.yaml"),
			outputPath:  filepath.Join("..", "otelcol", "config", "collector", "agent_config.yaml"),
		},
		{
			basePath:    filepath.Join("config", "agent_config.yaml"),
			overlayPath: filepath.Join("config", "agent_to_gateway_overlay.yaml"),
			outputPath:  filepath.Join("..", "otelcol", "config", "collector", "agent_to_gateway_config.yaml"),
		},
	}

	for _, overlay := range configs {
		mergedConfig := koanf.New(".")
		if err := mergedConfig.Load(file.Provider(overlay.basePath), yaml.Parser()); err != nil {
			fmt.Fprintf(os.Stderr, "failed to load base config %q: %v\n", overlay.basePath, err)
			os.Exit(1)
		}

		if err := mergedConfig.Load(file.Provider(overlay.overlayPath), yaml.Parser()); err != nil {
			fmt.Fprintf(os.Stderr, "failed to merge overlay config %q: %v\n", overlay.overlayPath, err)
			os.Exit(1)
		}

		mergedContents, err := mergedConfig.Marshal(yaml.Parser())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal merged config for %q: %v\n", overlay.outputPath, err)
			os.Exit(1)
		}

		if err := os.WriteFile(overlay.outputPath, mergedContents, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write merged config %q: %v\n", overlay.outputPath, err)
			os.Exit(1)
		}
	}
}
