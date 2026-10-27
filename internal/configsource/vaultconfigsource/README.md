# Vault Config Source

| Status | |
| --- | --- |
| Stability | Alpha |
| Distributions | [Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector) |

Use the [Vault](https://www.vaultproject.io/) config source to retrieve data from
Vault and inject it into your collector configuration. It supports
[dynamic secrets](https://www.vaultproject.io/),
[Key/Value V1 lease hints](https://www.vaultproject.io/docs/secrets/kv/kv-v1), and
[Key/Value V2 metadata polling](https://www.vaultproject.io/docs/secrets/kv/kv-v2).

## Configuration

Under `config_sources:`, use `vault:` or `vault/<name>:` to create a Vault config
source. `endpoint`, `path`, and one authentication method are required.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `endpoint` | Yes | None | Vault server URL. `VAULT_ADDR` can be referenced explicitly with `${env:VAULT_ADDR}`. |
| `path` | Yes | None | Vault path to the secret. |
| `auth` | Yes | None | Authentication settings. Set exactly one of `auth.token`, `auth.iam`, or `auth.gcp`. |
| `poll_interval` | No | `1m` | Interval for checking changes to non-dynamic KV v2 secrets. Has no effect on dynamic secrets, V1 leases, or non-dynamic secrets without KV v2 metadata. Must be positive. |

### Authentication options

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `auth.token` | For token authentication | None | Nonempty Vault token. `VAULT_TOKEN` can be referenced explicitly with `${env:VAULT_TOKEN}`. |
| `auth.iam` | For AWS IAM authentication | None | AWS IAM login settings; see below. |
| `auth.gcp` | For GCP authentication | None | GCP login settings; see below. |

For `auth.iam`, the following settings are available. AWS credentials can also
come from the standard AWS credential chain.

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `aws_access_key_id` | No | AWS credential chain | AWS access key ID. |
| `aws_secret_access_key` | No | AWS credential chain | AWS secret access key. |
| `aws_security_token` | No | AWS credential chain | AWS session token for temporary credentials. |
| `header_value` | No | Empty | Value of the `X-Vault-AWS-IAM-Server-ID` header. |
| `mount` | No | `aws` | Path where the AWS authentication method is mounted. |
| `role` | No | Empty | Vault role for the login request. |

For `auth.gcp`, the following settings are available:

| Option | Required | Default | Description |
| --- | --- | --- | --- |
| `role` | Yes | None | Vault role for the login request. |
| `mount` | No | `gcp` | Path where the GCP authentication method is mounted. |
| `credentials` | No | Unset | GCP credentials as a JSON string. Supplying them directly is not recommended. |
| `jwt_exp` | No | `15m` | Generated JWT lifetime; must be less than one hour. |
| `service_account` | No | `client_email` from `credentials`, when provided | Service account used to generate a JWT. |
| `project` | No | `project_id` from `credentials`, when provided | Project of the service account authenticating to Vault. |

### Example configuration

This example creates two named sources for different paths on the same Vault
server:

```yaml
config_sources:
  vault/kv:
    endpoint: ${env:VAULT_ADDR}
    path: secret/data/kv
    auth:
      token: ${env:VAULT_TOKEN}
  vault/db:
    endpoint: ${env:VAULT_ADDR}
    path: database/creds/collector_role
    auth:
      token: ${env:VAULT_TOKEN}

components:
  component_using_vault_kv:
    username: ${vault/kv:data.user}
    password: ${vault/kv:data.password}
  component_using_vault_db:
    username: ${vault/db:username}
    password: ${vault/db:password}
```

Here, `components` and the component names stand for valid component sections
and settings. With the Key/Value V2 engine, data is nested under `data` and
`metadata` maps; use a dot to select a nested value, such as `data.username`.
