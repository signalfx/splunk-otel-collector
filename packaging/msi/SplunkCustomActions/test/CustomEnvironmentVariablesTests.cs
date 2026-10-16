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

using WixToolset.Dtf.WindowsInstaller;

public class CustomEnvironmentVariablesTests
{
    [Fact]
    public void ParsesMultipleVariables()
    {
        var environment = CustomEnvironmentVariables.Parse("KEY1=VALUE1,KEY2=VALUE2");
        Assert.Equal(2, environment.Count);
        Assert.Equal("VALUE1", environment["KEY1"]);
        Assert.Equal("VALUE2", environment["KEY2"]);
    }

    [Theory]
    [InlineData("KEY=a%2Cb%2cc", "a,b,c")]
    [InlineData("KEY=a=b=c", "a=b=c")]
    [InlineData("KEY=a%3Db%3Bc%25%20%22%5B%5D", "a=b;c% \"[]")]
    [InlineData("KEY=%252C", "%2C")]
    [InlineData("KEY=a+b", "a+b")]
    [InlineData("KEY=%E2%9C%93", "✓")]
    [InlineData("KEY=", "")]
    [InlineData("KEY=  value  ", "  value  ")]
    public void DecodesValuesOnce(string input, string expected)
    {
        Assert.Equal(expected, CustomEnvironmentVariables.Parse(input)["KEY"]);
    }

    [Theory]
    [InlineData(null)]
    [InlineData("")]
    public void UnsetPropertyAddsNothing(string? input)
    {
        Assert.Empty(CustomEnvironmentVariables.Parse(input));
    }

    [Fact]
    public void LastDuplicateWinsIgnoringCase()
    {
        var environment = CustomEnvironmentVariables.Parse("KEY=first,key=last");
        Assert.Single(environment);
        Assert.Equal("last", environment["KEY"]);
    }

    [Theory]
    [InlineData("KEY")]
    [InlineData("=value")]
    [InlineData("%20=value")]
    [InlineData("KEY=value,unencoded-comma")]
    [InlineData("KEY=value,")]
    [InlineData(",KEY=value")]
    [InlineData("KEY%3DNAME=value")]
    [InlineData("KEY%00=value")]
    [InlineData("KEY=value%00")]
    public void RejectsInvalidEntries(string input)
    {
        var exception = Assert.Throws<ArgumentException>(() => CustomEnvironmentVariables.Parse(input));
        Assert.Contains("SPLUNK_ENVIRONMENT_VARIABLES", exception.Message);
    }

    [Fact]
    public void ErrorDoesNotDiscloseValues()
    {
        var exception = Assert.Throws<ArgumentException>(() =>
            CustomEnvironmentVariables.Parse("TOKEN=secret,invalid-entry"));
        Assert.DoesNotContain("secret", exception.Message);
    }

    [Fact]
    public void DeferredActionDataPreservesSpecialCharacters()
    {
        var data = CustomActions.BuildOptionalConfigurationData(name =>
            name == "SPLUNK_ENVIRONMENT_VARIABLES" ? "KEY=a%3Bb%3B%3Bc%3Dd,EMPTY=" : "");

        var transported = new CustomActionData(data.ToString());
        Assert.Equal("a;b;;c=d", transported["KEY"]);
        Assert.Equal("", transported["EMPTY"]);
    }

    [Fact]
    public void CustomVariablesOverrideOptionalPropertiesIgnoringCase()
    {
        var properties = new Dictionary<string, string>
        {
            ["GODEBUG"] = "original",
            ["SPLUNK_GATEWAY_URL"] = "http://gateway:4317",
            ["SPLUNK_ENVIRONMENT_VARIABLES"] = "godebug=custom%2Cvalue,GOMEMLIMIT=,CUSTOM=  "
        };
        var data = CustomActions.BuildOptionalConfigurationData(name =>
            properties.TryGetValue(name, out var value) ? value : "");
        var transported = new CustomActionData(data.ToString());

        Assert.Equal(4, transported.Count);
        Assert.Equal("custom,value", transported["GODEBUG"]);
        Assert.Equal("http://gateway:4317", transported["SPLUNK_GATEWAY_URL"]);
        Assert.Equal("", transported["GOMEMLIMIT"]);
        Assert.Equal("  ", transported["CUSTOM"]);
        Assert.False(transported.ContainsKey("SPLUNK_ENVIRONMENT_VARIABLES"));
    }
}
