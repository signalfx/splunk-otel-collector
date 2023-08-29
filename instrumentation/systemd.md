# Splunk OpenTelemetry Zero Configuration Auto Instrumentation for Systemd

The `splunk-otel-auto-instrumentation` Debian/RPM package provides examples of drop-in file(s) that can be copied to the
host's `systemd` configuration directory to activate/configure Auto Instrumentation agent(s) for supported applications
running as `systemd` services by defining default environment variables.

## Manual Systemd Configuration

> `systemd` supports many options, methods, and paths for configuring environment variables at the system level or for
> individual services, and are not limited to the examples below. Before making any changes, it is recommended to
> consult the documentation specific to your Linux distribution or service, and check the existing configuration of the
> system or individual services for potential conflicts. For general details about `systemd`, see the
> [`systemd` man page](https://www.freedesktop.org/software/systemd/man/index.html).

### Java Auto Instrumentation ###

See the [Advanced Configuration Guide](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
for details about supported options and defaults for the Java agent. These options can be configured via environment
variables or their corresponding system properties.

#### Configuration Priority ####
 
The Java agent can consume configuration options from one or more of the following sources (ordered from highest to
lowest priority):
1. Java system properties (`-D` flags) passed directly to the agent. For example,
     ```shell
     JAVA_TOOL_OPTIONS="-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     ```
2. Environment variables
3. Property files

See [Configuring the agent](
https://opentelemetry.io/docs/instrumentation/java/automatic/agent-config/#configuring-the-agent) for more
information.

#### Quick Start ####

The [`/usr/lib/splunk-instrumentation/examples/systemd/00-splunk-otel-javaagent.conf`](
packaging/fpm/examples/systemd/00-splunk-otel-javaagent.conf) example drop-in file defines the following environment variable:

- `JAVA_TOOL_OPTIONS=-javaagent:/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar`

To activate the Java agent and its default configuration for ***all*** Java applications running as `systemd` services,
copy this file to the host's `systemd` configuration directory, e.g. `/usr/lib/systemd/system.conf.d/`, and
reboot the system or run the following commands to restart the applicable services for the changes to take effect
(requires `root` privileges):
  ```shell
  $ systemctl daemon-reload
  $ systemctl restart <service-name>   # replace "<service-name>" and run for each applicable service
  ```

#### Configuration ####

To configure the Java agent, add/modify/override [supported environment variables](
https://docs.splunk.com/Observability/gdi/get-data-in/application/java/configuration/advanced-java-otel-configuration.html)
within `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf` (requires `root` privileges):

1. **Option A**: Add/Update `DefaultEnvironment` within `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`
   for the desired environment variables. For example:
     ```shell
     $ cat <<EOH > /usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     DefaultEnvironment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option B**: Create/Modify a higher-priority drop-in file for ***all*** services to add or override the environment
   variables defined in `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For example:
     ```shell
     $ cat <<EOH >> /usr/lib/systemd/system.conf.d/99-my-custom-env-vars.conf
     [Manager]
     DefaultEnvironment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     DefaultEnvironment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
   **Option C**: Create/Modify a higher-priority drop-in file for a ***specific*** service to add or override the
   environment variables defined in `/usr/lib/systemd/system.conf.d/00-splunk-otel-javaagent.conf`. For
   example:
     ```shell
     $ cat <<EOH >> /usr/lib/systemd/system/my-service.d/99-my-custom-env-vars.conf
     [Service]
     Environment="JAVA_TOOL_OPTIONS=-javaagent:/my/custom/splunk-otel-javaagent.jar -Dotel.service.name=my-service"
     Environment="OTEL_RESOURCE_ATTRIBUTES=myattrubute1=value1,myattribute2=value2"
     Environment="SPLUNK_PROFILER_ENABLED=true"
     EOH
     ```
2. After any configuration changes, reboot the system or run the following commands to restart the applicable services
   for the changes to take effect:
     ```shell
     $ systemctl daemon-reload
     $ systemctl restart <service-name>   # replace "<service-name>" and run for each applicable service
     ```
