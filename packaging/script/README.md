# OTel scripting

This folder contains the packaging of a solution to script with OpenTelemetry, based on the OpenTelemetry Python SDK.

## Building

Run the following make targets:

`$> make rpm-package`: make the RPM package

`$> make deb-package`: make the Debian package

`$> make tar-package`: make a simple tar.gz archive of the package.

## Testing

You can test each artifact created similarly to how they were built:

`$> make rpm-test`

`$> make deb-test`

`$> make tar-test`


