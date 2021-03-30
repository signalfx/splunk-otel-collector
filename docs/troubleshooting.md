# Troubleshooting

Start by reviewing the [OpenTelemetry Collector troubleshooting
documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/troubleshooting.md).

## Gathering Support Information

If you are unable to determine why something is not working then you can [email
support](mailto:signalfx-support@splunk.com). When opening a support request,
it is important to include as much information about the issue as possible
including:

- What did you try to do?
- What happened?
- What did you expect to happen?
- Have you found any workaround?
- How impactful is the issue?
- How can we produce the issue?

End-to-end architecture information is helpful including:

- What is generating the data?
- Where was the data configured to go to?
- What format was the data sent in?
- How is the next hop configured?
- Where is the data configured to go from here?
- What format was the data sent in?
- Any dns/firewall/networking/proxy information to be aware of?

In addition, it is important to gather support information including:

- Configuration file
  - Kubernetes: `kubectl get configmap my-configmap -o yaml >my-configmap.yaml`
  - Linux: `/etc/otel/collector`
- Logs and ideally debug logs
  - Docker: `docker logs my-container >my-container.log`
  - Journald: `journalctl -u my-service >my-service.log`
  - Kubernetes:
      `kubectl describe pod my-pod`
      `kubectl logs my-pod otel-collector >my-pod-otel.log`
      `kubectl logs my-pod fluentd >my-pod-fluentd.log`

Support bundle scripts are provided to make it easier to collect information:

- Linux (if installer script was used): `/etc/otel/collector/splunk-support-bundle.sh`

## Linux Installer

If either the splunk-otel-collector or td-agent services are not properly
installed and configured:

- Ensure the OS [is supported](getting-started/linux-installer.md#linux-installer-script)
- Ensure the OS has systemd installed
- Ensure not running in a containerized environment (for non-production
  environments see [this
  post](https://developers.redhat.com/blog/2014/05/05/running-systemd-within-docker-container/)
  for a workaround)
- Check installation logs for more details

## HTTP Error Codes

- 401 (UNAUTHORIZED): Configured access token or realm is incorrect
- 404 (NOT FOUND): Likely configuration parameter is wrong like endpoint or path
  (e.g. /v1/log); possible network/firewall/port issue
- 429 (TOO MANY REQUESTS): Org is not provisioned for the amount of traffic
  being sent; reduce traffic or request increase in capacity
- 503 (SERVICE UNAVAILABLE): If using the Log Observer, this is the same as 429
  (because that is how HECv1 responds). Otherwise, check the status page.

## Log Collection

### Is the source generating logs?

```bash
$ tail -f /var/log/foo.log
$ journalctl -u my-service.service -f
```

### Is Fluentd configured properly?

- Is td-agent running? (`systemctl status td-agent`)
- If you changed the configuration did you restart fluentd? (`systemctl restart td-agent`)
- Check `fluentd.conf` and `conf.d/\*`; ensure `@label @SPLUNK` is added to
  every source otherwise logs are not collected!
- Manual configuration may be required to collect logs off the source. Add
  configuration files to in the `conf.d` directory as needed.
- Enable debug logging in `fluentd.conf` (`log_level debug`), restart td-agent
  (`systemctl restart td-agent`), and check `/var/log/td-agent/td-agent.log`
- While every attempt is made to properly configure permissions, it is
  possible td-agent does not have the permission required to collect logs.
  Debug logging should indicate this issue.
- It is possible the `<parser>` section configuration is not matching the log events.
- This means things are working (requires debug logging enabled): `2021-03-17
  02:14:44 +0000 [debug]: #0 connect new socket`

### Is OTelCol configured properly?

- Check
  [zpages](https://github.com/open-telemetry/opentelemetry-collector/blob/main/extension/zpagesextension)
  for samples (`http://localhost:55679/debug/tracez`); may require `endpoint`
  configuration
- Enable [logging
  exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)
  and check logs (`journalctl -u splunk-otel-collector.service -f`)
- Review the [Collector troubleshooting
  documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/troubleshooting.md).

### Sending synthetic data

You can manually generate logs if needed. By default, Fluentd should monitor
journald and `/var/log/syslog.log` for events.

> Note: Properly structured syslog may be required for Fluentd to properly pick
> up the log line

```bash
$ echo "2021-03-17 02:14:44 +0000 [debug]: test" >>/var/log/syslog.log
$ echo "2021-03-17 02:14:44 +0000 [debug]: test" | systemctl-cat
```

## Trace Collection

### Sending synthetic data

How can you test the Collector is able to receive spans without instrumenting
an application?

By default, the Collector enables the Zipkin receiver, which is capable of
receiving trace data over JSON. Zipkin provides some example data
[here](https://github.com/openzipkin/zipkin/tree/master/zipkin-lens/testdata).
As a result, you can test by running something like the following:

```bash
$ curl -OL https://raw.githubusercontent.com/openzipkin/zipkin/master/zipkin-lens/testdata/yelp.json
$ curl -X POST localhost:9411/api/v2/spans -H'Content-Type: application/json' -d @yelp.json
```

> NOTE: Update `localhost` as appropriate to reach the Collector.

No response means the request was successfully sent. You can also pass `-v` to
the `curl` command to confirm.
