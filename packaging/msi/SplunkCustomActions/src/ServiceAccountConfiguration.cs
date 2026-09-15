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

using Microsoft.Win32;
using System.ComponentModel;
using System.IO;
using System.Runtime.InteropServices;
using System.Security.AccessControl;
using System.Security.Principal;

internal enum ServiceAccountType
{
    Preserve,
    LocalSystem,
    Virtual,
}

internal static class ServiceAccountConfiguration
{
    internal const string ServiceName = "splunk-otel-collector";
    internal const string VirtualAccountName = @"NT SERVICE\splunk-otel-collector";
    internal const string LocalSystemAccountName = "LocalSystem";

    private const uint ScManagerConnect = 0x0001;
    private const uint ServiceChangeConfig = 0x0002;
    private const uint ServiceNoChange = 0xffffffff;
    private const uint ServiceConfigServiceSidInfo = 5;
    private const uint ServiceSidTypeUnrestricted = 1;
    private const uint PolicyCreateAccount = 0x0010;
    private const uint PolicyLookupNames = 0x0800;
    private const int ErrorMemberInAlias = 1378;
    private const int ErrorNoSuchAlias = 1376;
    private const int NerrGroupNotFound = 2220;

    internal static ServiceAccountType ParseAccountType(string? accountType)
    {
        switch (accountType ?? string.Empty)
        {
            case "":
                return ServiceAccountType.Preserve;
            case "virtual":
                return ServiceAccountType.Virtual;
            case "localsystem":
                return ServiceAccountType.LocalSystem;
            default:
                throw new ArgumentException("SPLUNK_SERVICE_ACCOUNT_TYPE must be empty, 'virtual', or 'localsystem'.", nameof(accountType));
        }
    }

    internal static string SelectAccount(string? configuredAccountType, string? existingAccount)
    {
        string existingAccountName = existingAccount ?? string.Empty;
        switch (ParseAccountType(configuredAccountType))
        {
            case ServiceAccountType.Virtual:
                return VirtualAccountName;
            case ServiceAccountType.LocalSystem:
                return LocalSystemAccountName;
            case ServiceAccountType.Preserve:
                if (string.IsNullOrWhiteSpace(existingAccountName))
                {
                    return VirtualAccountName;
                }

                if (IsLocalSystem(existingAccountName))
                {
                    return LocalSystemAccountName;
                }

                if (string.Equals(existingAccountName, VirtualAccountName, StringComparison.OrdinalIgnoreCase))
                {
                    return VirtualAccountName;
                }

                throw new InvalidOperationException(
                    $"The existing service account '{existingAccountName}' cannot be preserved because Windows does not expose its password. " +
                    "Set SPLUNK_SERVICE_ACCOUNT_TYPE to 'virtual' or 'localsystem' explicitly.");
            default:
                throw new InvalidOperationException("Unknown service account type.");
        }
    }

    internal static bool IsLocalSystem(string accountName) =>
        string.Equals(accountName, LocalSystemAccountName, StringComparison.OrdinalIgnoreCase) ||
        string.Equals(accountName, @"NT AUTHORITY\SYSTEM", StringComparison.OrdinalIgnoreCase);

    internal static string GetExistingServiceAccount()
    {
        using (RegistryKey? serviceKey = Registry.LocalMachine.OpenSubKey($@"SYSTEM\CurrentControlSet\Services\{ServiceName}"))
        {
            return serviceKey?.GetValue("ObjectName") as string ?? string.Empty;
        }
    }

    internal static string LocalSystemRecommendation =>
        "The Splunk OpenTelemetry Collector service remains configured as LocalSystem. " +
        "To use the dedicated service account, run: msiexec.exe /i <path-to-splunk-otel-collector.msi> /qn SPLUNK_SERVICE_ACCOUNT_TYPE=virtual";

    internal static void Configure(string selectedAccount, string programDataPath, Action<string> log)
    {
        ConfigureServiceAccount(selectedAccount);

        if (!string.Equals(selectedAccount, VirtualAccountName, StringComparison.OrdinalIgnoreCase))
        {
            log($"Info: Splunk OpenTelemetry Collector service account is {selectedAccount}.");
            return;
        }

        SetUnrestrictedServiceSid();
        AddToLocalGroup("Performance Monitor Users", log);
        AddToLocalGroup("Event Log Readers", log);

        SecurityIdentifier serviceSid = ResolveVirtualAccountSid();
        GrantSecurityEventLogPrivilege(serviceSid, log);
        GrantFileAccess(
            Path.Combine(programDataPath, "Splunk", "OpenTelemetry Collector"),
            serviceSid,
            FileSystemRights.ReadAndExecute,
            true);

        string fileStoragePath = Path.Combine(programDataPath, "Splunk", "OpenTelemetry Collector", "FileStorage");
        GrantFileAccess(fileStoragePath, serviceSid, FileSystemRights.Modify, true);
        GrantFileAccess(Path.Combine(fileStoragePath, "Temp"), serviceSid, FileSystemRights.Modify, true);

        GrantDefaultFileLogAccess(serviceSid, log);
        log($"Info: Configured dedicated service account {VirtualAccountName}.");
    }

    private static void ConfigureServiceAccount(string accountName)
    {
        string? password = string.Equals(accountName, VirtualAccountName, StringComparison.OrdinalIgnoreCase)
            ? null
            : string.Empty;
        IntPtr scm = OpenSCManager(null, null, ScManagerConnect);
        if (scm == IntPtr.Zero)
        {
            throw new Win32Exception(Marshal.GetLastWin32Error(), "Failed to open the Service Control Manager.");
        }

        try
        {
            IntPtr service = OpenService(scm, ServiceName, ServiceChangeConfig);
            if (service == IntPtr.Zero)
            {
                throw new Win32Exception(Marshal.GetLastWin32Error(), $"Failed to open the {ServiceName} service.");
            }

            try
            {
                if (!ChangeServiceConfig(
                    service,
                    ServiceNoChange,
                    ServiceNoChange,
                    ServiceNoChange,
                    null,
                    null,
                    IntPtr.Zero,
                    null,
                    accountName,
                    password,
                    null))
                {
                    throw new Win32Exception(Marshal.GetLastWin32Error(), $"Failed to configure {ServiceName} to run as {accountName}.");
                }
            }
            finally
            {
                CloseServiceHandle(service);
            }
        }
        finally
        {
            CloseServiceHandle(scm);
        }
    }

    private static void SetUnrestrictedServiceSid()
    {
        IntPtr scm = OpenSCManager(null, null, ScManagerConnect);
        if (scm == IntPtr.Zero)
        {
            throw new Win32Exception(Marshal.GetLastWin32Error(), "Failed to open the Service Control Manager.");
        }

        try
        {
            IntPtr service = OpenService(scm, ServiceName, ServiceChangeConfig);
            if (service == IntPtr.Zero)
            {
                throw new Win32Exception(Marshal.GetLastWin32Error(), $"Failed to open the {ServiceName} service.");
            }

            try
            {
                ServiceSidInfo sidInfo = new ServiceSidInfo { ServiceSidType = ServiceSidTypeUnrestricted };
                IntPtr sidInfoPointer = Marshal.AllocHGlobal(Marshal.SizeOf(typeof(ServiceSidInfo)));
                try
                {
                    Marshal.StructureToPtr(sidInfo, sidInfoPointer, false);
                    if (!ChangeServiceConfig2(service, ServiceConfigServiceSidInfo, sidInfoPointer))
                    {
                        throw new Win32Exception(Marshal.GetLastWin32Error(), $"Failed to configure an unrestricted service SID for {ServiceName}.");
                    }
                }
                finally
                {
                    Marshal.FreeHGlobal(sidInfoPointer);
                }
            }
            finally
            {
                CloseServiceHandle(service);
            }
        }
        finally
        {
            CloseServiceHandle(scm);
        }
    }

    private static void AddToLocalGroup(string groupName, Action<string> log)
    {
        LocalGroupMembersInfo3 member = new LocalGroupMembersInfo3 { DomainAndName = VirtualAccountName };
        int result = NetLocalGroupAddMembers(null, groupName, 3, ref member, 1);
        switch (result)
        {
            case 0:
                log($"Info: Added {VirtualAccountName} to {groupName}.");
                return;
            case ErrorMemberInAlias:
                log($"Info: {VirtualAccountName} is already a member of {groupName}.");
                return;
            case ErrorNoSuchAlias:
            case NerrGroupNotFound:
                log($"Warning: The local {groupName} group is unavailable. This is expected on a domain controller; grant the required receiver permissions explicitly.");
                return;
            default:
                throw new Win32Exception(result, $"Failed to add {VirtualAccountName} to {groupName}.");
        }
    }

    private static SecurityIdentifier ResolveVirtualAccountSid()
    {
        return (SecurityIdentifier)new NTAccount(VirtualAccountName).Translate(typeof(SecurityIdentifier));
    }

    private static void GrantSecurityEventLogPrivilege(SecurityIdentifier serviceSid, Action<string> log)
    {
        LsaObjectAttributes objectAttributes = new LsaObjectAttributes
        {
            Length = Marshal.SizeOf(typeof(LsaObjectAttributes)),
        };
        uint status = LsaOpenPolicy(IntPtr.Zero, ref objectAttributes, PolicyCreateAccount | PolicyLookupNames, out IntPtr policy);
        if (status != 0)
        {
            throw new Win32Exception((int)LsaNtStatusToWinError(status), "Failed to open the local security policy.");
        }

        try
        {
            byte[] serviceSidBytes = new byte[serviceSid.BinaryLength];
            serviceSid.GetBinaryForm(serviceSidBytes, 0);
            IntPtr privilegeName = Marshal.StringToHGlobalUni("SeSecurityPrivilege");
            try
            {
                LsaUnicodeString right = new LsaUnicodeString
                {
                    Buffer = privilegeName,
                    Length = (ushort)("SeSecurityPrivilege".Length * sizeof(char)),
                    MaximumLength = (ushort)(("SeSecurityPrivilege".Length + 1) * sizeof(char)),
                };
                status = LsaAddAccountRights(policy, serviceSidBytes, new[] { right }, 1);
                if (status != 0)
                {
                    throw new Win32Exception((int)LsaNtStatusToWinError(status),
                        $"Failed to grant SeSecurityPrivilege to {VirtualAccountName}.");
                }
            }
            finally
            {
                Marshal.FreeHGlobal(privilegeName);
            }
        }
        finally
        {
            LsaClose(policy);
        }

        log($"Info: Granted SeSecurityPrivilege to {VirtualAccountName} for the Security event log receiver.");
    }

    private static void GrantDefaultFileLogAccess(SecurityIdentifier serviceSid, Action<string> log)
    {
        string windowsDirectory = Environment.GetFolderPath(Environment.SpecialFolder.Windows);
        GrantExistingFileAccess(Path.Combine(windowsDirectory, "System32", "DHCP"), serviceSid, true, log);
        GrantExistingFileAccess(Path.Combine(windowsDirectory, "WindowsUpdate.log"), serviceSid, false, log);
        GrantExistingFileAccess(Path.Combine(windowsDirectory, "debug", "netlogon.log"), serviceSid, false, log);
        GrantExistingFileAccess(Path.Combine(windowsDirectory, "System32", "LogFiles", "Firewall"), serviceSid, true, log);
    }

    private static void GrantExistingFileAccess(string path, SecurityIdentifier serviceSid, bool isDirectory, Action<string> log)
    {
        if ((isDirectory && !Directory.Exists(path)) || (!isDirectory && !File.Exists(path)))
        {
            log($"Warning: Default file log path '{path}' does not exist; grant {VirtualAccountName} read access before enabling it.");
            return;
        }

        GrantFileAccess(path, serviceSid, FileSystemRights.ReadAndExecute, isDirectory);
    }

    private static void GrantFileAccess(string path, SecurityIdentifier serviceSid, FileSystemRights rights, bool isDirectory)
    {
        if (isDirectory)
        {
            Directory.CreateDirectory(path);
            DirectoryInfo directory = new DirectoryInfo(path);
            DirectorySecurity security = directory.GetAccessControl();
            security.AddAccessRule(new FileSystemAccessRule(
                serviceSid,
                rights,
                InheritanceFlags.ContainerInherit | InheritanceFlags.ObjectInherit,
                PropagationFlags.None,
                AccessControlType.Allow));
            directory.SetAccessControl(security);
            return;
        }

        FileInfo file = new FileInfo(path);
        FileSecurity fileSecurity = file.GetAccessControl();
        fileSecurity.AddAccessRule(new FileSystemAccessRule(serviceSid, rights, AccessControlType.Allow));
        file.SetAccessControl(fileSecurity);
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct ServiceSidInfo
    {
        public uint ServiceSidType;
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    private struct LocalGroupMembersInfo3
    {
        [MarshalAs(UnmanagedType.LPWStr)]
        public string DomainAndName;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct LsaObjectAttributes
    {
        public int Length;
        public IntPtr RootDirectory;
        public IntPtr ObjectName;
        public uint Attributes;
        public IntPtr SecurityDescriptor;
        public IntPtr SecurityQualityOfService;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct LsaUnicodeString
    {
        public ushort Length;
        public ushort MaximumLength;
        public IntPtr Buffer;
    }

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    private static extern IntPtr OpenSCManager(string? machineName, string? databaseName, uint desiredAccess);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    private static extern IntPtr OpenService(IntPtr scm, string serviceName, uint desiredAccess);

    [DllImport("advapi32.dll", SetLastError = true, CharSet = CharSet.Unicode)]
    private static extern bool ChangeServiceConfig(
        IntPtr service,
        uint serviceType,
        uint startType,
        uint errorControl,
        string? binaryPathName,
        string? loadOrderGroup,
        IntPtr tagId,
        string? dependencies,
        string? serviceStartName,
        string? password,
        string? displayName);

    [DllImport("advapi32.dll", SetLastError = true)]
    private static extern bool ChangeServiceConfig2(IntPtr service, uint infoLevel, IntPtr info);

    [DllImport("advapi32.dll", SetLastError = true)]
    private static extern bool CloseServiceHandle(IntPtr serviceHandle);

    [DllImport("Netapi32.dll", CharSet = CharSet.Unicode)]
    private static extern int NetLocalGroupAddMembers(
        string? serverName,
        string groupName,
        int level,
        ref LocalGroupMembersInfo3 buffer,
        int totalEntries);

    [DllImport("advapi32.dll")]
    private static extern uint LsaOpenPolicy(
        IntPtr systemName,
        ref LsaObjectAttributes objectAttributes,
        uint desiredAccess,
        out IntPtr policyHandle);

    [DllImport("advapi32.dll")]
    private static extern uint LsaAddAccountRights(
        IntPtr policyHandle,
        byte[] accountSid,
        [In] LsaUnicodeString[] userRights,
        uint countOfRights);

    [DllImport("advapi32.dll")]
    private static extern uint LsaNtStatusToWinError(uint status);

    [DllImport("advapi32.dll")]
    private static extern uint LsaClose(IntPtr objectHandle);
}
