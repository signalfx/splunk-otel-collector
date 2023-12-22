// Copyright Splunk, Inc.
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

//go:build integration

package tests

import (
	"path"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var sshd = testutils.NewContainer().WithContext(path.Join(".", "testdata", "server")).WithExposedPorts("2222:22").WithName("sshd-server").WillWaitForPorts("22")

func TestSSHCheckReceiverBadPassword(t *testing.T) {
	testutils.AssertAllMetricsReceived(t, "sshcheck_bad_password.yaml", "sshcheck_config_bad_password.yaml",
		[]testutils.Container{sshd}, nil)
}

func TestSSHCheckReceiverHappyPath(t *testing.T) {
	testutils.AssertAllMetricsReceived(t, "sshcheck_happy_path.yaml", "sshcheck_config_happy_path.yaml",
		[]testutils.Container{sshd}, nil)
}
