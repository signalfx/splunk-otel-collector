# Run the Collector with a dedicated Windows service account

The Splunk OpenTelemetry Collector can run as the Windows virtual service
account `NT SERVICE\splunk-otel-collector`. This passwordless account is
dedicated to the Collector service and limits its access compared with
LocalSystem.

## Account selection

For a new MSI installation, the dedicated account is selected by default. An
upgrade preserves the existing account when it is LocalSystem or the Collector
virtual service account. If an upgrade uses another account, select a supported
account explicitly because the installer cannot recover and preserve that
account's password.

To select an account explicitly with the PowerShell installer script, use
`-service_account_type`:

```powershell
.\install.ps1 -access_token "ACCESSTOKEN" -service_account_type virtual
```

Use `virtual` to run under `NT SERVICE\splunk-otel-collector`, or
`localsystem` to run under LocalSystem. The option applies to both new
installations and upgrades. For an MSI installation or upgrade, set the
equivalent public property to `virtual` or `localsystem`:

```powershell
msiexec.exe /i splunk-otel-collector.msi /qn SPLUNK_SERVICE_ACCOUNT_TYPE=virtual
```

If the property is omitted on an upgrade, the MSI preserves a supported
existing account. An upgrade from another account must specify `virtual` or
`localsystem` so the service can be migrated without requiring the old
password.

When an MSI upgrade preserves LocalSystem without an explicit selection, the
installer logs a recommendation to migrate to the dedicated account. The
PowerShell installer also prints the corresponding MSI command.

## Permissions for the dedicated account

When the MSI configures the virtual service account, it grants only the
permissions needed for the default Collector installation:

- Membership in the local `Performance Monitor Users` and `Event Log Readers`
  groups, for performance counter and event log collection.
- The `SeSecurityPrivilege` right, for reading the Security event log.
- Read and execute access to the Collector installation and configuration
  directories.
- Modify access to the Collector's file storage, its `Temp` subdirectory, and
  the OpAMP Supervisor state directory.
- Read access to existing default file log sources.

The MSI grants access to known default log paths that exist when it runs. If a
custom receiver reads additional files, grant the service account access to
those files and directories separately. On a domain controller, Windows may
not provide the local groups; the installer reports this as a warning.

To return an installed service to LocalSystem manually, run an elevated
PowerShell session:

```powershell
sc.exe config splunk-otel-collector obj= LocalSystem
Restart-Service splunk-otel-collector
```
