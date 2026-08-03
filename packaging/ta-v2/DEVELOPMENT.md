# Development Guide

The overall goal is that all resources to develop and test are either part of the
repository or can easily be obtained from public sources. This works well for
when the development is targeting Linux, however for Windows there are some
manual steps involved.

## Local Docker Testing

Use [`run-local-package-in-container.sh`](./run-local-package-in-container.sh)
on Linux to run the TA in the official `splunk/splunk` image. The mounted app
name defaults to `Splunk_TA_otel_linux_x86_64`; set `APP_NAME` to override it.
Set `ASSETS_DIR`, `CONTAINER_NAME`, or `SPLUNK_VERSION` to override the assets
directory, Docker container name, or Splunk image tag.

## Windows Docker Support

Unlike Linux, there is no official Docker image for Windows that includes Splunk.
Therefore, in order to create a Windows-based Splunk test environment, you
can use the following resources:

- [`Dockerfile.windows`](./Dockerfile.windows): A Windows image that installs
  the Splunk Universal Forwarder and prepares the environment by adding necessary
  folders and setting permissions. See comments in the Dockerfile for more details.
  The admin account is created with username `admin` and a password that must
  be supplied at build time with `--build-arg SPLUNK_ADMIN_PASSWORD=<password>`
  (the build fails if it is not set), so the REST API at `localhost:8089`
  (mapped by `run-local-package-in-container.ps1`) can be used to authenticate.
- [`run-local-package-in-container.ps1`](./run-local-package-in-container.ps1):
  This PowerShell script automates the testing of the Splunk Technical Add-on (TA)
  for OpenTelemetry Collector by spinning up a Windows-based Splunk Universal
  Forwarder container with the TA package mounted inside it. The script sets up
  directories for assets and logs, stops any existing test container, launches a
  new Docker container running splunk-uf-windows with the collector binaries
  mounted from a local assets directory to the appropriate Splunk apps location,
  and then monitors the container's splunkd.log file to verify that the
  Splunk_TA_otel application is successfully loaded and recorded.
  The mounted app name defaults to `Splunk_TA_otel_windows_x86_64`; set
  `APP_NAME` to override it. Set `ASSETS_DIR`, `CONTAINER_NAME`, or `IMAGE_TAG`
  to override the assets directory, Docker container name, or local
  `splunk-uf-windows` image tag.
- [`windows-container-start-and-wait.ps1`](./windows-container-start-and-wait.ps1):
  A PowerShell script that starts the Splunk Universal Forwarder service inside
  the running container and waits for it to be fully operational before proceeding
  with further actions.
