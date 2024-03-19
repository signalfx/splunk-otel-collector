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

using Microsoft.Deployment.WindowsInstaller;

public class CustomActions
{
    /// <summary>
    /// Custom action to check if the launch conditions are met.
    /// </summary>
    /// <param name="session">Carries information about the current installation session.</param>
    /// <returns>An action result that indicates success or failure of the check.</returns>
    [CustomAction]
    public static ActionResult CheckLaunchConditions(Session session)
    {
        // Check if it is uninstall, if so, skip any check.
        var removeValue = session["REMOVE"];
        session.Log("Info: REMOVE=" + removeValue);
        if (removeValue == "ALL")
        {
            return ActionResult.Success;
        }

        // If SPLUNK_SETUP_COLLECTOR_MODE is not one of the expected values, fail the check.
        var collectorMode = session["SPLUNK_SETUP_COLLECTOR_MODE"] ?? string.Empty;
        session.Log("Info: SPLUNK_SETUP_COLLECTOR_MODE=" + collectorMode);
        if (collectorMode != "agent" && collectorMode != "gateway")
        {
            LogAndShowError(session, "SPLUNK_SETUP_COLLECTOR_MODE must be either 'agent' or 'gateway'.");
            return ActionResult.Failure;
        }

        return ActionResult.Success;
    }

    /// <summary>
    /// Helper to log and show error messages if the installation fails.
    /// When running in silent mode no dialog is shown.
    /// <summary>
    /// <param name="session">Carries information about the current installation session.</param>
    /// <param name="message">The error message to log and show.</param>
    private static void LogAndShowError(Session session, string message)
    {
        session.Log("Error: " + message);
        if (session["UILevel"] != InstallUIOptions.Silent.ToString()) // 2 is the value for INSTALLUILEVEL_NONE
        {
            session.Message(InstallMessage.Error, new Record { FormatString = message });
        }
    }
}
