# Development Guide

The overall goal is that all resources to develop and test are either part of the
repository or can easily be obtained from public sources. This works well for
when the development is targeting Linux, however for Windows there are some
manual steps involved.

## Windows Docker Support

Unlike Linux, there is no official Docker image for Windows that includes Splunk.
Therefore, in order to create a Windows-based Splunk test environment, you
can use the following resources:

- `[Dockerfile.windows]`(./Dockerfile.windows): A Windows image that installs
  the Splunk Universal Forwarder and prepares the environment by adding necessary
  folders and setting permissions.
- `[run-local-package-in-container.ps1]`(./run-local-package-in-container.ps1):
  A PowerShell script that builds the Docker image using the provided Dockerfile
  and runs a container from that image. It maps the local `out/distribution`
  folder to the appropriate directory inside the container, allowing you to test
  the Technical Add-on.
- `[windows-container-start-and-wait.ps1]`(./windows-container-start-and-wait.ps1):
  A PowerShell script that starts the Splunk Universal Forwarder service inside
  the running container and waits for it to be fully operational before proceeding
  with further actions.
