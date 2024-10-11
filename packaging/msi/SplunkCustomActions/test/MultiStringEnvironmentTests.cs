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

public class MultiStringEnvironmentTests
{
    private const string TestSubKey = @"Software\SplunkOpenTelemetryCollectorTest";
    private const string TestKey = @"HKEY_CURRENT_USER\" + TestSubKey;
    private const string TestValueName = "TestMultiString";

    [Fact]
    public void GetEnvironmentValue()
    {
        try
        {
            var initialEnvironment = new string[]
            {
                "z_key0=value0",
                "a_key1=value1"
            };
            Registry.SetValue(TestKey, TestValueName, initialEnvironment.ToArray(), RegistryValueKind.MultiString);

            using (var multiStringEnvironment = new MultiStringEnvironment(RegistryHive.CurrentUser, TestSubKey, TestValueName))
            {
                var actualEnvironment = multiStringEnvironment.GetEnvironmentValue();
                actualEnvironment.Should().BeEquivalentTo(initialEnvironment);
            }
        }
        finally
        {
            DeleteTestSubKey();
        }
    }

    [Fact]
    public void AddingOptionalConfigurations()
    {
        try
        {
            var initialEnvironment = new string[]
            {
                "key0=value0",
                "key1=value1"
            };
            Registry.SetValue(TestKey, TestValueName, initialEnvironment.ToArray(), RegistryValueKind.MultiString);

            using (var multiStringEnvironment = new MultiStringEnvironment(RegistryHive.CurrentUser, TestSubKey, TestValueName))
            {
                multiStringEnvironment.AddEnvironmentVariables(new Dictionary<string, string>
                {
                    { "key2", "value2" },
                    { "key3", "value3" }
                });
            }

            var expectedEnvironment = new string[]
            {
                "key0=value0",
                "key1=value1",
                "key2=value2",
                "key3=value3"
            };

            Registry.GetValue(TestKey, TestValueName, Array.Empty<string>()).Should().BeEquivalentTo(expectedEnvironment);
        }
        finally
        {
            DeleteTestSubKey();
        }
    }

    [Fact]
    public void DefaultConstructorShouldFail()
    {
        Action action = () => new MultiStringEnvironment();
        action
            .Should()
            .Throw<InvalidOperationException>()
            .WithMessage("Failed to open the registry sub key: SYSTEM\\CurrentControlSet\\Services\\splunk-otel-collector");
    }

    private void DeleteTestSubKey()
    {
        Registry.CurrentUser.DeleteSubKeyTree(TestSubKey);
    }
}
