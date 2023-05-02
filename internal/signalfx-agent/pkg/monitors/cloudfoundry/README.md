# cloudfoundry-firehose-nozzle

## Developer Resources

- <https://github.com/cf-platform-eng/firehose-nozzle-v2/tree/master/gateway>
- <https://github.com/cloudfoundry/loggregator/blob/master/docs/rlp_gateway.md>
- <https://github.com/cloudfoundry/go-loggregator/blob/master/rlp_gateway_client.go>

## Tanzu Application Service Setup

### Create a new TAS environment

1. Get access to [Pivotal Partners Slack](https://pivotalpartners.slack.com/archives/C42PWTRR9)
1. Create a new TAS environment via: <https://self-service.isv.ci/>

### Configure TAS for monitoring

Required CLI tools:

- [`hammer`](https://github.com/pivotal/hammer)
- [`bosh`](https://github.com/cloudfoundry/bosh-cli)
- [`cf`](https://github.com/cloudfoundry/cli) (most probably v6)
  - [nozzle-plugin](https://github.com/cloudfoundry-community/firehose-plugin)
- [`uaac`](https://github.com/cloudfoundry/cf-uaac)
- [`om`](https://github.com/pivotal-cf/om)
- [`jq`](https://stedolan.github.io/jq/)

1. Download the hammer config from <https://self-service.isv.ci> and name it like your environement and export a variable

    ```sh
    export TAS_JSON=<path to the downloaded JSON>
    ```

2. Create a new space and configure it as a default target space

    ```sh
    hammer -t $TAS_JSON cf-login
    cf create-space test-space && cf target -s test-space
    ```

3. Deploy a sample application:

    ```sh
    hammer -t $TAS_JSON cf-login
    git clone https://github.com/cloudfoundry-samples/test-app && cd test-app && cf push && cd .. && rm -rf test-app && cf apps
    ```

4. Create a UAA user with the proper permissions to access the RLP Gateway:

    ```sh
    eval "$(hammer -t $TAS_JSON om)"
    UAA_CREDS=$(om credentials -p cf -c .uaa.identity_client_credentials -t json | jq '.password' -r)
    TAS_SYS_DOMAIN=$(jq '.sys_domain' -r $TAS_JSON)
    uaac target https://uaa.$TAS_SYS_DOMAIN
    uaac token client get identity -s $UAA_CREDS
    NOZZLE_SECRET=$(openssl rand -hex 12)
    uaac client add my-v2-nozzle --name signalfx-nozzle --secret $NOZZLE_SECRET --authorized_grant_types client_credentials,refresh_token --authorities logs.admin
    echo "signalfx-nozzle client secret: $NOZZLE_SECRET"
    ```

### Configure SignalFx Smart Agent monitor

Example config:

```yaml
---
signalFxAccessToken: <signalfx token>
intervalSeconds: 10
logging:
  level: debug
monitors:
  - type: cloudfoundry-firehose-nozzle
    rlpGatewayUrl: https://log-stream.sys.<TAS environement name>.cf-app.com
    rlpGatewaySkipVerify: true
    uaaUser: my-v2-nozzle
    uaaPassword: <signalfx-nozzle client secret>
    uaaUrl: https://uaa.sys.<TAS environement name>.cf-app.com
    uaaSkipVerify: true
# Required: What format to send data in
writer:
  traceExportFormat: sapm
```

More: [docs/monitors/cloudfoundry-firehose-nozzle.md](../../../docs/monitors/cloudfoundry-firehose-nozzle.md)
