# Splunk OpenTelemetry Connector Heroku Buildpack

A Heroku buildpack to install and run the Splunk OpenTelemetry Connector on a
Dyno and send data to Splunk Observability Cloud.

> :construction: This project is currently in **BETA**

## Getting Started

[Install the Heroku CLI, login, and create an
app](https://devcenter.heroku.com/articles/heroku-cli). Add and configure the
buildpack:

```
# cd into the Heroku project directory
# WARNING: running `heroku` command outside of project directories
#          will result in unexpected behavior
#cd test
#git init
#heroku apps:create splunk-example

# Configure Heroku App to expose Dyno metadata
# This metadata is required by the Splunk OpenTelemetry Connector to
# set global dimensions such as `app_name`, `app_id` and `dyno_id`.
# See [here](https://devcenter.heroku.com/articles/dyno-metadata) for more information.
heroku labs:enable runtime-dyno-metadata

# Add buildpack for Splunk OpenTelemetry Connector
heroku buildpacks:add https://github.com/signalfx/splunk-otel-collector-heroku.git#<BUILDPACK_VERSION>
# Required for test application
#heroku buildpacks:add heroku/nodejs

# Setup required environment variables
heroku config:set SPLUNK_ACCESS_TOKEN=<YOUR_ACCESS_TOKEN>
heroku config:set SPLUNK_REALM=<YOUR_REALM>

# Optionally define custom configuration file
#heroku config:set SPLUNK_OTEL_CONFIG=/app/test/config.yaml

# If these buildpacks are being added to an existing project,
# create an empty commit prior to deploying the app
git commit --allow-empty -m "empty commit"

# Deploy your app
git push heroku main

# Check logs
#heroku logs -a splunk-example --tail
```

## Advanced Configuration

Use the following environment variables to configure this buildpack

| Environment Variable      | Required | Default                                             | Description                                                                     |
| ----------------------    | -------- | -------                                             | -------------------------------------------------------------------------       |
| `SPLUNK_REALM`            | Yes      | `us0`                                               | Your Splunk realm.                                                              |
| `SPLUNK_TOKEN`            | Yes      |                                                     | Your Splunk access token.                                                       |
| `SPLUNK_API_URL`          | No       | `https://api.SPLUNK_REALM.signalfx.com`             | The Splunk API base URL.                                                        |
| `SPLUNK_CONFIG`           | No       | `/app/config.yaml`                                  | The configuration to use. `/app/.splunk/config.yaml` used if default not found. |
| `SPLUNK_INGEST_URL`       | No       | `https://ingest.SPLUNK_REALM.signalfx.com`          | The Splunk Infrastructure Monitoring base URL.                                  |
| `SPLUNK_LOG_FILE`         | No       | `/dev/stdout`                                       | Specify location of agent logs. If not specified, logs will go to stdout.       |
| `SPLUNK_MEMORY_TOTAL_MIB` | No       | `512`                                               | Total available memory to agent.                                                |
| `SPLUNK_OTEL_VERSION`     | No       | `latest`                                            | Version of Splunk OTel Connector to use. Defaults to latest.                    |
| `SPLUNK_TRACE_URL`        | No       | `https://ingest.SPLUNK_REALM.signalfx.com/v2/trace` | The Splunk APM base URL.                                                        |
