// Copyright Splunk, Inc.
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

//go:build testutils

package manifests

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

type one struct {
	Thing        string
	AnotherThing string
}

func (o one) Render(tb testing.TB) string {
	tmplt := `---
thing: "{{.Thing}}"
anotherThing: "{{.AnotherThing}}"
`
	man := Manifest[one](tmplt)
	return man.Render(o, tb)
}

func TestManifest(t *testing.T) {
	man := Manifest[one]("{{.Thing}}")
	manifest := man.Render(one{Thing: "thing"}, t)
	require.Equal(t, "thing", manifest)

	man = Manifest[one]("{{ .AnotherThing | upper }}")
	manifest = man.Render(one{AnotherThing: "another.thing"}, t)
	require.Equal(t, "ANOTHER.THING", manifest)
}

func TestRenderAll(t *testing.T) {
	first := one{}
	second := one{Thing: "some", AnotherThing: "thing"}
	third := one{AnotherThing: "another.thing"}
	require.Equal(t, `---
thing: ""
anotherThing: ""
---
thing: "some"
anotherThing: "thing"
---
thing: ""
anotherThing: "another.thing"
`, RenderAll(t, first, second, third))
}

func TestK8sYaml(t *testing.T) {
	type Thing struct {
		Container corev1.Container
	}

	man := Manifest[Thing](`container:
{{ .Container | toYaml | indent 2 }}
`)
	manifest := man.Render(Thing{Container: corev1.Container{
		Name:       "container-name",
		Image:      "container-image",
		Command:    []string{"container", "command"},
		Args:       []string{"arg.one", "arg.two"},
		WorkingDir: "/working/dir",
		Ports:      []corev1.ContainerPort{{Name: "a.port", HostPort: 123, ContainerPort: 234, Protocol: corev1.ProtocolUDP, HostIP: "1.2.3.4"}},
	}}, t)
	require.Equal(t, `container:
  args:
  - arg.one
  - arg.two
  command:
  - container
  - command
  image: container-image
  name: container-name
  ports:
  - containerPort: 234
    hostIP: 1.2.3.4
    hostPort: 123
    name: a.port
    protocol: UDP
  resources: {}
  workingDir: /working/dir
`, manifest)
}
