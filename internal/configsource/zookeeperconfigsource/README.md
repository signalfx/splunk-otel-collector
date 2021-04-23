# Zookeeper Config Source (Alpha)

Use the [Zookeeper](https://zookeeper.apache.org/) config source to retrieve data from
Zookeeper and inject it into your collector configuration.

## Configuration

Under the `config_sources:` use `zookeeper:` or `zookeeper/<name>:` to create a 
Zookeeper config source. The following parameters are available to customize
Zookeeper config sources:

```yaml
config_sources:
  zookeeper:
    # endpoint is the Zookeeper server addresses. Config source will try to connect to
    # these endpoints to access an Zookeeper cluster.
    endpoints: [http://localhost:2181]
    # timeout sets the amount of time for which a session is considered valid after
    # losing connection to a server. Within the session timeout it's possible to 
    # reestablish a connection to a different server and keep the same session.
    timeout: 10s
```

If multiple paths are needed create different instances of the config source, example:

```yaml
config_sources:
    # Assuming that the environment variables ZOOKEEPER_ADDR is the defined and the 
    # different secrets are on the same server but at different paths.
    zookeeper:
      endpoints: [$ZOOKEEPER_ADDR]
    zookeeper/another_cluster:
      endpoints: [$ZOOKEEPER_2_ADDR]
      timeout: 15s

# Both Zookeeper config sources can be used via their full name. Hypothetical example:
components:
  component_using_zookeeper:
    token: $zookeeper:/data/token

  component_using_zookeeper_another_cluster:
    token: $zookeeper/another_cluster:/data/token
```
