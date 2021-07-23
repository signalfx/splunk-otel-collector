# Example Heroku App

Simple NodeJS application based on [SignalFx
NodeJS](https://github.com/signalfx/signalfx-nodejs). Emits metrics to local
Splunk OpenTelemetry Connector. See Procfile for run command. To run locally:

```
npm install signalfx
node send_metrics.js
```
