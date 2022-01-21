# linux-autoinstrumentation

This directory contains functionality to auto instrument Java programs that run on a Linux host.

The installation process will be done automatically in a later release, but for now, you can build a `.so` file for
Linux with `make all`, then place the `.so` file on the host you want to auto instrument and put the full path to the
`.so` file in a text file at the location `/etc/ld.so.preload`. Make sure you have a
[Splunk OTel Java jar](https://github.com/signalfx/splunk-otel-java) at the location
`/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`. Once done, the `.so` will set the environment variable
`JAVA_TOOL_OPTIONS` to the path of your agent jar every time `java` is run on the host, causing all Java executables
on the host to be auto instrumented.
