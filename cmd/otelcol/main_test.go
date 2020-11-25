package main

import (
	"os"
	"strconv"
	"testing"
)

func TestContains(t *testing.T) {
	testArgs := [][]string{
		[]string{"cmd", "--test=foo"},
		[]string{"cmd", "--test", "foo"},
	}
	for _, v := range testArgs {
		result := contains(v, "--test")
		if !(result) {
			t.Errorf("Expected true got false while testing %v", v)
		}
	}
	testArgs = [][]string{
		[]string{"cmd", "--test-fail", "foo"},
		[]string{"cmd", "--test-fail=--test"},
	}
	for _, v := range testArgs {
		result := contains(v, "--test")
		if result {
			t.Errorf("Expected false got true while testing %v", v)
		}
	}
}

func TestUseMemorySizeFromEnvVar(t *testing.T) {
	testArgs := [][]string{
		[]string{"", "0"},
	}
	for _, v := range testArgs {
		n, _ := strconv.Atoi(v[1])
		os.Setenv(ballastEnvVarName, v[0])
		useMemorySizeFromEnvVar(n)
		result := contains(os.Args, "--mem-ballast-size-mib")
		if result {
			t.Errorf("Expected false got true while testing %v", v)
		}
	}
	testArgs = [][]string{
		[]string{"10", "0"},
		[]string{"", "10"},
	}
	for _, v := range testArgs {
		n, _ := strconv.Atoi(v[1])
		os.Setenv(ballastEnvVarName, v[0])
		useMemorySizeFromEnvVar(n)
		result := contains(os.Args, "--mem-ballast-size-mib")
		if !(result) {
			t.Errorf("Expected true got false while testing %v", v)
		}
	}
}

func TestUseConfigFromEnvVar(t *testing.T) {
	result := contains(os.Args, "--config")
	if result {
		t.Error("Expected false got true while looking for --config")
	}
	os.Setenv(configEnvVarName, "../../"+defaultLocalSAPMNonLinuxConfig)
	useConfigFromEnvVar()
	result = contains(os.Args, "--config")
	if !(result) {
		t.Error("Expected true got false while looking for --config")
	}
	os.Unsetenv(configEnvVarName)
}

func TestCheckMemorySettingMiBFromEnvVar(t *testing.T) {
	testArgs := [][]string{
		[]string{"", "10", "0"},
		[]string{"10", "0", "10"},
		[]string{"10", "100", "10"},
	}
	for _, v := range testArgs {
		n, _ := strconv.Atoi(v[1])
		r, _ := strconv.Atoi(v[2])
		os.Setenv(memLimitMiBEnvVarName, v[0])
		result := checkMemorySettingsMiBFromEnvVar(memLimitMiBEnvVarName, n)
		if result != r {
			t.Errorf("Expected %d but got %d", r, result)
		}
	}
	os.Unsetenv(memLimitMiBEnvVarName)
}

func HelperUseMemorySettingsFromEnvVar(limitEnv string, limit int, spikeEnv string, spike int, t *testing.T) {
	envVarVal, _ := strconv.Atoi(os.Getenv(limitEnv))
	if envVarVal != limit {
		t.Errorf("Expected %d but got %d", limit, envVarVal)
	}
	envVarVal, _ = strconv.Atoi(os.Getenv(spikeEnv))
	if envVarVal != spike {
		t.Errorf("Expected %d but got %d", spike, envVarVal)
	}
	os.Unsetenv(limitEnv)
	os.Unsetenv(spikeEnv)
}

func TestSetMemorySettingsToEnvVar(t *testing.T) {
	setMemorySettingsToEnvVar(10, memLimitMiBEnvVarName, 5, memSpikeMiBEnvVarName)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 10, memSpikeMiBEnvVarName, 5, t)
}
func TestUseMemorySettingsMiBFromEnvVar(t *testing.T) {
	useMemorySettingsMiBFromEnvVar(100)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 90, memSpikeMiBEnvVarName, 25, t)
	useMemorySettingsMiBFromEnvVar(30000)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 27952, memSpikeMiBEnvVarName, 2048, t)

	setMemorySettingsToEnvVar(10, memLimitMiBEnvVarName, 5, memSpikeMiBEnvVarName)
	useMemorySettingsMiBFromEnvVar(0)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 10, memSpikeMiBEnvVarName, 5, t)
	useMemorySettingsMiBFromEnvVar(100)
	HelperUseMemorySettingsFromEnvVar(memLimitMiBEnvVarName, 90, memSpikeMiBEnvVarName, 25, t)
}

func TestCheckMemorySettingsPercentageFromEnvVar(t *testing.T) {
	result := checkMemorySettingsPercentageFromEnvVar(memLimitEnvVarName, defaultMemoryLimitPercentage)
	if result != defaultMemoryLimitPercentage {
		t.Errorf("Expected %d but got %d", defaultMemoryLimitPercentage, result)
	}
	setMemorySettingsToEnvVar(10, memLimitEnvVarName, 5, memSpikeEnvVarName)
	result = checkMemorySettingsMiBFromEnvVar(memLimitEnvVarName, 10)
	if result != 10 {
		t.Errorf("Expected %d but got %d", 10, result)
	}
	os.Unsetenv(memLimitEnvVarName)
	os.Unsetenv(memSpikeEnvVarName)
}

func TestUseMemorySettingsPercentageFromEnvVar(t *testing.T) {
	useMemorySettingsPercentageFromEnvVar()
	HelperUseMemorySettingsFromEnvVar(memLimitEnvVarName, 90, memSpikeEnvVarName, 25, t)
	setMemorySettingsToEnvVar(10, memLimitEnvVarName, 5, memSpikeEnvVarName)
	HelperUseMemorySettingsFromEnvVar(memLimitEnvVarName, 10, memSpikeEnvVarName, 5, t)
}
