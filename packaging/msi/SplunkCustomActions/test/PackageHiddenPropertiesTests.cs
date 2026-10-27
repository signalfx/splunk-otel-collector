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

using System.Xml.Linq;

public class PackageHiddenPropertiesTests
{
    private static readonly XNamespace Wix = "http://wixtoolset.org/schemas/v4/wxs";

    private static Dictionary<string, XElement> LoadProperties()
    {
        var path = Path.Combine(AppContext.BaseDirectory, "splunk-otel-collector.wxs");
        return XDocument.Load(path)
            .Descendants(Wix + "Property")
            .Where(property => property.Attribute("Id") != null)
            .ToDictionary(property => property.Attribute("Id")!.Value);
    }

    private static bool IsHidden(XElement property) => (string?)property.Attribute("Hidden") == "yes";

    [Fact]
    public void AllTokenPropertiesAreHidden()
    {
        var tokens = LoadProperties().Where(p => p.Key.Contains("TOKEN")).ToList();
        Assert.NotEmpty(tokens);

        var exposed = tokens.Where(p => !IsHidden(p.Value)).Select(p => p.Key).ToList();
        Assert.True(exposed.Count == 0, "Properties holding tokens must be Hidden: " + string.Join(", ", exposed));
    }

    [Theory]
    [InlineData("SPLUNK_ACCESS_TOKEN")]
    [InlineData("SPLUNK_HEC_TOKEN")]
    [InlineData("SPLUNK_PLATFORM_TOKEN")]
    [InlineData("SPLUNK_ENVIRONMENT_VARIABLES")]
    [InlineData("AddOptionalConfigurations")]
    public void SensitivePropertiesAreHidden(string name)
    {
        var properties = LoadProperties();
        Assert.True(properties.ContainsKey(name), name + " is not declared in the package.");
        Assert.True(IsHidden(properties[name]), name + " must be Hidden to keep its value out of the install log.");
    }
}
