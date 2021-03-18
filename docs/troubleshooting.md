# Troubleshooting

## Linux Installer

If either the splunk-otel-collector or td-agent services are not properly
installed and configured:

- Ensure the OS [is supported](getting-started/linux-installer.md#linux-installer-script)
- Ensure the OS has systemd installed
- Ensure not running in a containerized environment (for non-production environments see [this post](https://developers.redhat.com/blog/2014/05/05/running-systemd-within-docker-container/) for a workaround)
- Check installation logs for more details

## HTTP Error Codes

- 401 (UNAUTHORIZED): Configured access token or realm is incorrect
- 404 (NOT FOUND): Likely configuration parameter is wrong like endpoint or URI
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

You can manually generate logs if needed

> Note: Properly structured syslog may be required for Fluentd to properly pick
> up the log line

```bash
$ echo "2021-03-17 02:14:44 +0000 [debug]: test" >>/var/log/syslog.log
$ echo "2021-03-17 02:14:44 +0000 [debug]: test" | systemctl-cat
```

### Is Fluentd configured properly?

- Is td-agent running? (`systemctl status td-agent`)
- Check `fluentd.conf` and `conf.d/\*`; ensure `@label @SPLUNK` is added to
  every source otherwise logs are not collected!
- Manual configuration may be required to collect logs off the source. Add
  configuration files to in the `conf.d` directory as needed.
- Enable debug logging in `fluentd.conf` (`log_level debug`), restart td-agent
  (`systemctl restart td-agent`), and check `/var/log/td-agent/td-agent.log`
- While every attempt is made to properly configure permissions, it is
  possible td-agent does not have the permission required to collect logs.
  Debug logging should indicate this issue.
- This means things are working (requires debug logging enabled): `2021-03-17
  02:14:44 +0000 [debug]: #0 connect new socket`

### Is OTelCol configured properly?

- Check [zpages](https://github.com/open-telemetry/opentelemetry-collector/blob/main/extension/zpagesextension) for samples (`http://localhost:55679/debug/tracez`); may require
  `endpoint` configuration
- Enable [logging exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter) and check logs (`journalctl -u
  splunk-otel-collector.service -f`)
- Review the [Collector troubleshooting
  documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/master/docs/troubleshooting.md).
