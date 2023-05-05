// Copyright  Splunk, Inc.
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

//go:build !windows

package bundle

import (
	"embed"
)

// BundledFS is the in-executable filesystem that contains all bundled discovery config.d components.
//
// If you are bootstrapping bundle_gen.go or the `discoverybundler` cmd without any rendered files in bundle.d,
// comment out the below embed directives before installing to prevent "no matching files found"
// build errors.
//
//go:embed bundle.d/extensions/*.discovery.yaml
//go:embed bundle.d/receivers/*.discovery.yaml
var BundledFS embed.FS
