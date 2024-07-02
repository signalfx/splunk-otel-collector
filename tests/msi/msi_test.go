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

//go:build windows

package msi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func TestMSI(t *testing.T) {
	msiInstallerPath := os.Getenv("MSI_COLLECTOR_PATH")
	require.NotEmpty(t, msiInstallerPath, "MSI_COLLECTOR_PATH environment variable is not set")
	_, err := os.Stat(msiInstallerPath)
	require.NoError(t, err)
	if strings.Contains(msiInstallerPath, " ") {
		msiInstallerPath = "\"" + msiInstallerPath + "\""
	}

	tests := []struct {
		name          string
		msiProperties map[string]string
	}{
		{
			name: "default-install",
			msiProperties: map[string]string{
				"SPLUNK_ACCESS_TOKEN": "test\"token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installLogFile := filepath.Join(os.TempDir(), "install.log")
			// Build the MSI installation arguments and include the msiProperties map.
			args := []string{"/i", msiInstallerPath, "/qn", "/l*v", installLogFile}
			for key, value := range tt.msiProperties {
				// Escape the key and value if they contain spaces or quotes
				// See https://learn.microsoft.com/en-us/windows/win32/msi/command-line-options
				if strings.Contains(value, "\"") || strings.Contains(value, " ") {
					value = strings.ReplaceAll(value, "\"", "\"\"")
					value = "\"" + value + "\""
				// 	t.Logf("Escaped key: %s, value: [%s]", key, value)
				}
				args = append(args, key+"="+value)
			}

			// Run the MSI installer
			installCmd := exec.Command("msiexec")
			// msiexec is one of the noticeable exceptions about how to format the parameters,
			// see https://pkg.go.dev/os/exec#Command, so we need to join the args manually.
			cmdLine := strings.Join(args, " ")
			installCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec "+cmdLine}
			err := installCmd.Run()
			if err != nil {
				logText, _ := os.ReadFile(installLogFile)
				t.Log(string(logText))
			}
			t.Logf("Install command: %s", installCmd.SysProcAttr.CmdLine)
			require.NoError(t, err, "Failed to install the MSI: %v\nArgs: %v", err, args)

			defer func() {
				// Uninstall the MSI
				uninstallCmd := exec.Command("msiexec")
				uninstallCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "msiexec /x " + msiInstallerPath + " /qn"}
				err := uninstallCmd.Run()
				t.Logf("Uninstall command: %s", uninstallCmd.SysProcAttr.CmdLine)
				require.NoError(t, err, "Failed to uninstall the MSI: %v", err)
			}()

			// Verify the service
			scm, err := mgr.Connect()
			require.NoError(t, err)
			defer scm.Disconnect()

			service, err := scm.OpenService("splunk-otel-collector")
			require.NoError(t, err)
			defer service.Close()

			err = service.Start()
			require.NoError(t, err)
			defer func() {
				_, err = service.Control(svc.Stop)
				require.NoError(t, err)

				require.Eventually(t, func() bool {
					status, err := service.Query()
					require.NoError(t, err)
					return status.State == svc.Stopped
				}, 10*time.Second, 500*time.Millisecond, "Failed to stop the service")
			}()

			// Wait for the service to reach the running state
			require.Eventually(t, func() bool {
				status, err := service.Query()
				require.NoError(t, err)
				return status.State == svc.Running
			}, 10*time.Second, 500*time.Millisecond, "Failed to start the service")
		})
	}
}
