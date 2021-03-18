# Troubleshooting

## Linux Installer

If either the splunk-otel-collector or td-agent services are not properly
installed and configured:

- Ensure the OS is supported
- Ensure the OS has systemd installed
- Ensure pid1 can be accessed (custom configuration required for containerized
  environments)
- Check installation logs for more details

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
