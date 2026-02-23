[Splunk_TA_OTel_Collector://<name>]

splunk_access_token = <value>
* Required, token used to send data to Splunk Observability
* Default =

splunk_config = <value>
* Optional, config file that will be used by the Splunk_TA_OTel_Collector
* Default = $SPLUNK_HOME/etc/apps/Splunk_TA_OTel_Collector/configs/agent_config.yaml

splunk_realm = <value>
* Optional, Splunk Observability realm to which data will be sent to
* Default = us0

splunk_api_url = <value>
* Optional, specifies the Splunk Observability API endpoint
* Default = https://api.$SPLUNK_REALM.signalfx.com

splunk_ingest_url = <value>
* Optional, specifies the Splunk Observability ingest endpoint
* Default = https://ingest.$SPLUNK_REALM.signalfx.com

splunk_listen_interface = <value>
* Optional, address for the listening interfaces opened by the Splunk_TA_OTel_Collector
* Default = localhost

splunk_collector_log_level = <value>
* Optional, specifies the log level to be used by the Splunk_TA_OTel_Collector
* Default = error
