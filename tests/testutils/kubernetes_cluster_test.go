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
//go:build integration

package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateCluster(t *testing.T) {
	tc := NewTestcase(t)
	cluster := CreateCluster(tc)
	defer cluster.Delete()

	namespace := &apiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}}
	ns, err := cluster.Clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	require.NoError(t, err)
	require.Equal(t, "test-namespace", ns.Name)
}
