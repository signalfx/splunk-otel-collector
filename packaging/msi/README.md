# Custom service environment variables

Set the MSI property `SPLUNK_ENVIRONMENT_VARIABLES` to a comma-separated list of
`KEY=VALUE` entries to configure arbitrary environment variables for the Collector
service:

```powershell
msiexec.exe /i splunk-otel-collector.msi /qn SPLUNK_ACCESS_TOKEN=your-token SPLUNK_ENVIRONMENT_VARIABLES="OTEL_RESOURCE_ATTRIBUTES=deployment.environment.name%3Dproduction%2Cservice.namespace%3Dpayments,HTTPS_PROXY=http://proxy:8080"
```

Commas inside values **must** be percent encoded as `%2C`. Other special characters
can also be percent encoded, for example `%3D` for `=`, `%3B` for `;`, `%22` for `"`,
and `%25` for a literal `%`. Values are decoded once; `+` remains a literal plus.
An empty value (`KEY=`) is supported. Entries must have a nonempty name, and neither
names nor values may contain NUL characters.

Custom entries override predefined service environment variables with the same
name. Names are compared without regard to case, and the last duplicate entry
wins. These entries only set the service environment; they do not change other MSI
configuration decisions such as selecting a configuration file.
