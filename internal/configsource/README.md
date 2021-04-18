# Vault Config Source

Use the [Vault](https://www.vaultproject.io/) config source to retrieve data from
Vault and inject it into your collector configuration. It supports:

- [Dynamic secrets](https://www.vaultproject.io/);
- [Key/Value V1 lease hints](https://www.vaultproject.io/docs/secrets/kv/kv-v1);
- [Key/Value V2 metadata polling](https://www.vaultproject.io/docs/secrets/kv/kv-v2);

## Configuration

Under the `config_sources:` use `vault:` or `vault/<name>:` to create a Vault config
source Under the latter section, the following parameters are available to customize
Vault config sources:

```yaml
config_sources:
  vault:
    # endpoint is the Vault server address. It can also be provided by the standard Vault
    # environment variable VAULT_ADDR. This option takes priority over the environment
    # variable if provided.
    endpoint: http://localhost:8200
    # path is the Vault path to the secret location.
    path: secret/data/kv
    # token is used to access the Vault server. The standard Vault environment variable
    # VAULT_TOKEN can be used instead. This option takes priority over the environment
    # variable if provided.
    token: some_toke_value
    # is used only for non-dynamic V2 K/V secret stores. It is the interval in which the
    # config source will check for changes on the data on the given Vault path. Defaults
    # to 1 minute if not specified.
    poll_interval: 90s
```

If multiple paths are needed create different instances of the config source, example:

```yaml
config_sources:
    # Assuming that the environment variables VAULT_ADDR and VAULT_TOKEN are the defined
    # and the different secrets are on the same server but at different paths.
    vault/kv:
      path: secret/data/kv
    vault/db:
      path: database/creds/collector_role

# Both Vault config sources can be used via their full name. Hypothetical example:
components:
  component_using_vault_kv:
    # Example showing K/V V2, see note below about the '.' usage.
    username: $vault/kv:data.user
    password: $vault/kv:data.password

  component_using_vault_db:
    username: $vault/db:username
    password: $vault/db:password
```

*Note:* When using the Key/Value V2 secret engine, all data will be nested under a
separate data map within the secret, e.g. `data` and `metadata`, to access specific
keys specify the "map" and the "key" using a `.` as separator, eg: `data.username`.
