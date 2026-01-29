# Development Guide

The overall goal is that all resources to develop and test are either part of the
repository or can easily be obtained from public sources. This works well for
when the development is targeting Linux, however for Windows there are some
manual steps involved.

## Windows Docker Support

Unlike Linux, there is no official Docker image for Windows that includes Splunk.
Therefore, in order to create a Windows-based Splunk test environment, you
can use the following resources:

- [`Dockerfile.windows`](./Dockerfile.windows): A Windows image that installs
  the Splunk Universal Forwarder and prepares the environment by adding necessary
  folders and setting permissions. See comments in the Dockerfile for more details.
- [`run-local-package-in-container.ps1`](./run-local-package-in-container.ps1):
  This PowerShell script automates the testing of the Splunk Technical Add-on (TA)
  for OpenTelemetry Collector by spinning up a Windows-based Splunk Universal
  Forwarder container with the TA package mounted inside it. The script sets up
  directories for assets and logs, stops any existing test container, launches a
  new Docker container running splunk-uf-windows with the collector binaries
  mounted from a local assets directory to the appropriate Splunk apps location,
  and then monitors the container's splunkd.log file to verify that the
  Splunk_TA_OTel_Collector application is successfully loaded and recorded.
- [`windows-container-start-and-wait.ps1`](./windows-container-start-and-wait.ps1):
  A PowerShell script that starts the Splunk Universal Forwarder service inside
  the running container and waits for it to be fully operational before proceeding
  with further actions.
