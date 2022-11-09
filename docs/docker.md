# Docker Component Requirements

This distribution includes the [Docker container stats monitor](https://docs.splunk.com/Observability/gdi/docker/docker.html)
via the [Smart Agent Receiver](../pkg/receiver/smartagent/README.md) to provide the ability to report metrics from containers running on your system.
It also includes the [Docker Observer Extension](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/observer/dockerobserver) to enable dynamically
instantiating your receivers using the [Receiver Creator](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/receivercreator/README.md) as target service containers
are reported by the [Docker daemon](https://docs.docker.com/config/daemon/). Both of these components use a client
that requires establishing a connection with the daemon similar to using the `docker` cli. In order for this to occur, you'll likely
need to do one of the following depending on your Docker installation and Collector deployment method.

## Domain Socket

If your daemon is listening to a domain socket (e.g. `/var/run/docker.sock`) then the collector service or executable needs to be granted
appropriate permissions and access.

### Linux installation

For most manual or scripted Linux installations, the `splunk-otel-collector` user should be added to the `docker` or similar group as configured on your system:

```bash
$ usermod -aG docker splunk-otel-collector
```

### Docker image

When using the [`quay.io/signalfx/splunk-otel-collector`](https://quay.io/repository/signalfx/splunk-otel-collector) image, the default container user should be added to the required group as configured on your system, and the domain socket should be bind-mounted:

```bash
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro --group-add $(stat -c '%g' /var/run/docker.sock) quay.io/signalfx/splunk-otel-collector:latest <...>
# or if specifying the user:group directly
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro --user "splunk-otel-collector:$(stat -c '%g' /var/run/docker.sock)" quay.io/signalfx/splunk-otel-collector:latest <...>
```
