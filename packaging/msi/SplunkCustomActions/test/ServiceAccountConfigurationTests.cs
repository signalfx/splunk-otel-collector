// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

public class ServiceAccountConfigurationTests
{
    [Theory]
    [InlineData("", ServiceAccountType.Preserve)]
    [InlineData("virtual", ServiceAccountType.Virtual)]
    [InlineData("localsystem", ServiceAccountType.LocalSystem)]
    public void ParseAccountTypeAcceptsSupportedValues(string value, ServiceAccountType expected)
    {
        Assert.Equal(expected, ServiceAccountConfiguration.ParseAccountType(value));
    }

    [Fact]
    public void ParseAccountTypeRejectsUnsupportedValue()
    {
        Assert.Throws<ArgumentException>(() => ServiceAccountConfiguration.ParseAccountType("domain-user"));
    }

    [Theory]
    [InlineData("", "", ServiceAccountConfiguration.VirtualAccountName)]
    [InlineData("", "LocalSystem", ServiceAccountConfiguration.LocalSystemAccountName)]
    [InlineData("", "NT AUTHORITY\\SYSTEM", ServiceAccountConfiguration.LocalSystemAccountName)]
    [InlineData("", "NT SERVICE\\splunk-otel-collector", ServiceAccountConfiguration.VirtualAccountName)]
    [InlineData("virtual", "LocalSystem", ServiceAccountConfiguration.VirtualAccountName)]
    [InlineData("localsystem", "NT SERVICE\\splunk-otel-collector", ServiceAccountConfiguration.LocalSystemAccountName)]
    public void SelectAccountHandlesFreshInstallsUpgradesAndOverrides(string configuredType, string existingAccount, string expected)
    {
        Assert.Equal(expected, ServiceAccountConfiguration.SelectAccount(configuredType, existingAccount));
    }

    [Fact]
    public void SelectAccountRequiresAnExplicitChoiceForAnUnsupportedExistingAccount()
    {
        Assert.Throws<InvalidOperationException>(() => ServiceAccountConfiguration.SelectAccount("", @"CONTOSO\collector"));
    }

    [Fact]
    public void LocalSystemRecommendationHasTheExactMigrationCommand()
    {
        Assert.Contains(
            "msiexec.exe /i <path-to-splunk-otel-collector.msi> /qn SPLUNK_SERVICE_ACCOUNT_TYPE=virtual",
            ServiceAccountConfiguration.LocalSystemRecommendation);
    }
}
