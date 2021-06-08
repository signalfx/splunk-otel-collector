# End-to-end Tests

The tests in this directory are designed to help confirm that data reported
to Splunk IMM by the Collector is of expected content and form.  In order to
configure your tested Collector and any test clients to reach Splunk IMM, you will
need to provide credentials and endpoints in a `./testdata/secrets.yaml` file of the
form (as applicable):

```yaml
---
default_token: "<my SFx client token>"
api_token: "<my org token with API auth scope>"
ingest_token: "<my org token with Ingest auth scope>"
ingest_api_token: "<my org token with Ingest and API auth scopes>"
api_url: "<my API url (e.g. https://api.realm.signalfx.com)>"
ingest_url: "<my Ingest url (e.g. https://ingest.realm.signalfx.com)>"
signalflow_url: "<my SignalFlow url (e.g. wss://stream.realm.signalfx.com/v2/signalflow)>"
```

Feel free to add any helpful fields as necessary, and as more tests are added
a proper realm and credential fixtures or helpers should be broken out to avoid redundant
setup logic.  We should also later create common SFx api and signalflow clients that use the
specified `default_token` but at this time leave that responsibility to the tests.

Any `SPLUNK_`-prefixed environment variables set in your current or test environment will
be populated to any `Testcase.SplunkOtelCollector()` instances your tests stand up as
a url and secret passing mechanism to your Collector config.
A per-`Testcase` `SPLUNK_TEST_ID` UUIDv4 environment variable and field (`Testcase.ID`)
is available for unique metric and metadata items to query to better ensure test isolation.
