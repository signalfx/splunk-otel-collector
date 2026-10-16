# Telegraf Exec Monitor Deprecation

The Smart Agent `telegraf/exec` monitor is deprecated. Using it requires
explicitly enabling the `receiver.smartagentreceiver.TelegrafExec` feature gate:

```sh
--feature-gates=+receiver.smartagentreceiver.TelegrafExec
```

If a Collector configuration uses `telegraf/exec` while the gate is disabled,
configuration validation fails and reports how to enable the gate. When the
monitor starts with the gate enabled, it logs a deprecation notice. Other Smart
Agent monitors are unaffected.

## Timeline

- **Deprecation notice**: Current release
- **Planned removal**: December 2026
