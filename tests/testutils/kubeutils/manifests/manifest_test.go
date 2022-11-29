// Copyright Splunk, Inc.
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

package manifests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManifest(t *testing.T) {
	type One struct {
		Thing        string
		AnotherThing string
	}

	man := Manifest[One]("{{.Thing}}")
	manifest, err := man.Render(One{Thing: "thing"})
	require.NoError(t, err)
	require.Equal(t, "thing", manifest)

	man = Manifest[One]("{{ .AnotherThing | upper }}")
	manifest, err = man.Render(One{AnotherThing: "another.thing"})
	require.NoError(t, err)
	require.Equal(t, "ANOTHER.THING", manifest)
}
