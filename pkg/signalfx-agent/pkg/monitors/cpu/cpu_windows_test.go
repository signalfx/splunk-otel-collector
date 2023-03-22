//go:build windows
// +build windows

package cpu

import (
	"testing"
)

func Test_ntQuerySystemInformation(t *testing.T) {
	t.Run("ensure ntdll.dll and ntQuerySystemInformation are available", func(t *testing.T) {
		if ntdllLoadErr != nil {
			t.Errorf("NtDLL was not loaded. Make sure it's still a feature in this version of windows.")
		}
		if procNtQuerySystemInformationLoadErr != nil {
			t.Errorf("NtQuerySystemInformation proc could not be loaded.  Make sure it's still a feature in this version of windows.")
		}
		got, err := ntQuerySystemInformation()
		if err != nil {
			t.Errorf("ntQuerySystemInformation() error = %v", err)
			return
		}
		if len(got) < 1 {
			t.Errorf("no per core times returned")
			return
		}
	})
}
