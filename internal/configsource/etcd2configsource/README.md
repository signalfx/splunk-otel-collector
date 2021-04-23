# Etcd2 Config Source (Alpha)

Use the [Etcd2](https://etcd.io/docs/v2.3/) config source to retrieve data from
Etcd2 and inject it into your collector configuration.

## Configuration

Under the `config_sources:` use `etcd2:` or `etcd2/<name>:` to create a Etcd2 config
source. The following parameters are available to customize Etcd2 config sources:

```yaml
config_sources:
  etcd2:
    # endpoint is the Etcd2 server addresses. Config source will try to connect to
    # these endpoints to access an Etcd2 cluster.
    endpoints: [http://localhost:2379]
    # auth is a optional section used to indicate the authentication method to be used.
    # currently only username and password is supported.
    auth:
      # username is the etcd2 username used to identify the etcd2 user. 
      username: etcd2_username
      # password is password of the user specifying in the username field.
      password: etcd2_password 
```

If multiple paths are needed create different instances of the config source, example:

```yaml
config_sources:
    # Assuming that the environment variables ETCD2_ADDR, ETCD2_USERNAME and $ETCD_PASSWORD 
    # are the defined and the different secrets are on the same server but at different paths.
    etcd2:
      endpoints: [$ETCD2_ADDR]
    etcd2/withauth:
      endpoints: [$ETCD2_ADDR]
      auth:
        username: $ETCD2_USERNAME
        password: $ETCD2_PASSWORD

# Both Etcd2 config sources can be used via their full name. Hypothetical example:
components:
  component_using_etcd2:
    token: $etcd2:/data/token

  component_using_etcd2_withauth:
    token: $etcd2/withauth:/data/token
```