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
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestClusterRoleBinding(t *testing.T) {
	crb := ClusterRoleBinding{
		Name:               "some.cluster.role.binding",
		Namespace:          "some.namespace",
		ClusterRoleName:    "some.cluster.role",
		ServiceAccountName: "some.service.account",
	}
	manifest, err := crb.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: some.cluster.role.binding
  namespace: some.namespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: some.cluster.role
subjects:
- kind: ServiceAccount
  name: some.service.account
  namespace: some.namespace
`, manifest)
}

func TestClusterRoleBindingWithRoleRefAndSubjects(t *testing.T) {
	crb := ClusterRoleBinding{
		Name:            "some.cluster.role.binding",
		Namespace:       "some.namespace",
		ClusterRoleName: "should.be.ignored",
		RoleRef: &rbacv1.RoleRef{
			APIGroup: "some.api.group",
			Kind:     "some.role.ref.kind",
			Name:     "some.role.ref.name",
		},
		ServiceAccountName: "should.be.ignored",
		Subjects: []rbacv1.Subject{
			{
				Kind:      "some.subject.kind",
				APIGroup:  "another.api.group",
				Name:      "some.subject.name",
				Namespace: "some.subject.namespace",
			},
		},
	}
	manifest, err := crb.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: some.cluster.role.binding
  namespace: some.namespace
roleRef:
  apiGroup: some.api.group
  kind: some.role.ref.kind
  name: some.role.ref.name
subjects:
- apiGroup: another.api.group
  kind: some.subject.kind
  name: some.subject.name
  namespace: some.subject.namespace
`, manifest)
}

func TestEmptyClusterRoleBinding(t *testing.T) {
	crb := ClusterRoleBinding{}
	manifest, err := crb.Render()
	require.NoError(t, err)
	require.Equal(t,
		`---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
`, manifest)
}
