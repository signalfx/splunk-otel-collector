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

using Microsoft.Win32;

public class MultiStringEnvironment : IDisposable
{
    private readonly string _valueName;
    private readonly RegistryKey _subKey;

    public MultiStringEnvironment() :
        this(RegistryHive.LocalMachine, @"SYSTEM\CurrentControlSet\Services\splunk-otel-collector", "Environment")
    {
    }

    public MultiStringEnvironment(RegistryHive registryHive, string subKey, string valueName)
    {
        _valueName = valueName;
        using var key = RegistryKey.OpenBaseKey(registryHive, RegistryView.Registry64)
            ?? throw new InvalidOperationException($"Failed to open registry hive: {registryHive}");
        _subKey = key.OpenSubKey(subKey, writable: true)
            ?? throw new InvalidOperationException($"Failed to open the registry sub key: {subKey}");
    }

    public string[] GetEnvironmentValue()
    {
        return (string[])(_subKey.GetValue(_valueName) ?? Array.Empty<string>());
    }

    public void AddEnvironmentVariables(Dictionary<string, string> environmentVariables)
    {
        string[] existingEnvironmentVariables = GetEnvironmentValue();
        string[] newEnvironment = 
            [.. existingEnvironmentVariables, .. environmentVariables.Select(kvp => $"{kvp.Key}={kvp.Value}")];

        // Sort the environment variables to ensure that the order is consistent
        Array.Sort(newEnvironment, StringComparer.OrdinalIgnoreCase);

        _subKey.SetValue(_valueName, newEnvironment, RegistryValueKind.MultiString);
    }

    public void Dispose()
    {
        _subKey.Close();
    }
}
