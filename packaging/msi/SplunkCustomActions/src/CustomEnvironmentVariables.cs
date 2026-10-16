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

internal static class CustomEnvironmentVariables
{
    public static Dictionary<string, string> Parse(string? value)
    {
        var environment = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase);
        if (value == null || value.Length == 0)
        {
            return environment;
        }

        foreach (var entry in value.Split(','))
        {
            var separator = entry.IndexOf('=');
            if (separator <= 0)
            {
                throw InvalidEntry();
            }

            var name = Uri.UnescapeDataString(entry.Substring(0, separator));
            var variableValue = Uri.UnescapeDataString(entry.Substring(separator + 1));
            if (string.IsNullOrWhiteSpace(name) || name.IndexOfAny(['=', '\0']) >= 0 ||
                variableValue.IndexOf('\0') >= 0)
            {
                throw InvalidEntry();
            }

            environment[name] = variableValue;
        }

        return environment;
    }

    private static ArgumentException InvalidEntry() => new ArgumentException(
        "SPLUNK_ENVIRONMENT_VARIABLES must contain comma-separated KEY=VALUE entries with nonempty names. " +
        "Commas in values must be percent encoded as %2C; names cannot contain '=' and entries cannot contain NUL.");
}
