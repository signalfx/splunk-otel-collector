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

using WixToolset.Dtf.WindowsInstaller;

public class CustomActions
{
    [CustomAction]
    public static ActionResult DetectExistingServiceAccount(Session session)
    {
        if (session["REMOVE"] == "ALL")
        {
            return ActionResult.Success;
        }

        string existingAccount = session["SPLUNK_PREVIOUS_SERVICE_ACCOUNT"];
        if (string.IsNullOrWhiteSpace(existingAccount))
        {
            existingAccount = ServiceAccountConfiguration.GetExistingServiceAccount();
        }

        session["DETECTED_SERVICE_ACCOUNT"] = existingAccount;
        if (!string.IsNullOrWhiteSpace(existingAccount))
        {
            session.Log($"Info: Detected existing Splunk OpenTelemetry Collector service account: {existingAccount}");
            try
            {
                ServiceAccountConfiguration.SelectAccount(session["SPLUNK_SERVICE_ACCOUNT_TYPE"], existingAccount);
            }
            catch (InvalidOperationException exception)
            {
                LogAndShowError(session, exception.Message);
                return ActionResult.Failure;
            }
        }

        return ActionResult.Success;
    }

    [CustomAction]
    public static ActionResult ValidateServiceAccountType(Session session)
    {
        try
        {
            ServiceAccountConfiguration.ParseAccountType(session["SPLUNK_SERVICE_ACCOUNT_TYPE"]);
            return ActionResult.Success;
        }
        catch (ArgumentException exception)
        {
            LogAndShowError(session, exception.Message);
            return ActionResult.Failure;
        }
    }

    [CustomAction]
    public static ActionResult LogServiceAccountRecommendation(Session session)
    {
        if (string.IsNullOrWhiteSpace(session["SPLUNK_SERVICE_ACCOUNT_TYPE"]) &&
            ServiceAccountConfiguration.IsLocalSystem(session["DETECTED_SERVICE_ACCOUNT"]))
        {
            session.Log("Warning: " + ServiceAccountConfiguration.LocalSystemRecommendation);
        }

        return ActionResult.Success;
    }

    [CustomAction]
    public static ActionResult ConfigureServiceAccount(Session session)
    {
        try
        {
            string selectedAccount = ServiceAccountConfiguration.SelectAccount(
                session.CustomActionData["SPLUNK_SERVICE_ACCOUNT_TYPE"],
                session.CustomActionData["DETECTED_SERVICE_ACCOUNT"]);
            ServiceAccountConfiguration.Configure(
                selectedAccount,
                session.CustomActionData["PROGRAMDATA"],
                message => session.Log(message));
            return ActionResult.Success;
        }
        catch (Exception exception)
        {
            session.Log("Error: " + exception.Message);
            return ActionResult.Failure;
        }
    }

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
    /// Custom action to add optional configuration options to the environment at the service scope.
    /// </summary>
    /// <remarks>
    /// It is possible to add these as machine-wide environment variables, but, for consistency,
    /// this custom action adds them to the service environment.
    /// </remarks>
    /// <param name="session">Carries information about the current installation session.</param>
    /// <returns>Always success, in case of error the code will throw an exception and fails the install.</returns>
    [CustomAction]
    public static ActionResult AddOptionalConfigurations(Session session)
    {
        var optionalEnvironmentVariables = new Dictionary<string, string>();
        foreach (var kvp in session.CustomActionData)
        {
            if (!string.IsNullOrWhiteSpace(kvp.Value))
            {
                session.Log($"Info: Setting environment variable {kvp.Key}={kvp.Value}");
                optionalEnvironmentVariables[kvp.Key] = kvp.Value;
            }
        }

        if (optionalEnvironmentVariables.Count > 0)
        {
            using (var environment = new MultiStringEnvironment())
            {
                environment.AddEnvironmentVariables(optionalEnvironmentVariables);
            }
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
