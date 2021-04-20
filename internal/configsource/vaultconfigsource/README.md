# Vault Config Source (Alpha)

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
    # endpoint is the Vault server address. It is equivalent to the Vault tool
    # environment variable VAULT_ADDR.
    endpoint: http://localhost:8200
    # path is the Vault path to the secret location.
    path: secret/data/kv
    # poll_interval is used only for non-dynamic V2 K/V secret stores. It is
    # the interval in which the config source will check for changes on the
    # data on the given Vault path. Defaults to 1 minute if not specified.
    poll_interval: 90s
    # auth is a section used to indicate the authentication method to be used.
    # Exactly one method must be specified, it must be one of the following:
    # "token", "iam", or "gcp".
    auth:
      # token is used to access the Vault server. It is equivalent to the Vault tool
      # environment variable VAULT_TOKEN.
      token: some_toke_value
      # iam is used on AWS deployments to generate the required Vault token.
      # For details about each of the settings below, see
      # https://github.com/hashicorp/vault/blob/v1.1.0/builtin/credential/aws/cli.go#L148
      iam:
        aws_access_key_id: key_id
        aws_secret_access_key: access_key
        aws_security_token: security_token
        header_value: header_value
        mount: aws
        role: role
      # gcp is used on GCP deployments to generate the required Vault token.
      # For details about each of the settings below, see
      # https://github.com/hashicorp/vault-plugin-auth-gcp/blob/e1f6784b379d277038ca0661606aa8d23791e392/plugin/cli.go#L138
      gcp:
        role: role
        mount: gcp
        credentials: json_string # This setting is not recommended.
        jwp_ext: 10m
        service_account: some_account
        project: project_id
```

If multiple paths are needed create different instances of the config source, example:

```yaml
config_sources:
    # Assuming that the environment variables VAULT_ADDR and VAULT_TOKEN are the defined
    # and the different secrets are on the same server but at different paths.
    vault/kv:
      endpoint: $VAULT_ADDR
      path: secret/data/kv
      auth:
        token: $VAULT_TOKEN
    vault/db:
      endpoint: $VAULT_ADDR
      path: database/creds/collector_role
      auth:
        token: $VAULT_TOKEN

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
