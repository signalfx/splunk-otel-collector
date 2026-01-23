[Splunk_TA_OTel_Collector://<name>]

splunk_access_token = <value>
* Required, token used to send data to Splunk Observability
* Default =

splunk_api_url = <value>
* Optional, specifies the Splunk Observability API endpoint 
* Default = https://api.us0.signalfx.com

splunk_ingest_url = <value>
* Optional, specifies the Splunk Observability ingest endpoint
* Default = https://ingest.us0.signalfx.com

splunk_listen_interface = <value>
* Optional, address for the listening interfaces open by the Splunk_TA_OTel_Collector
* Default = localhost

splunk_realm = <value>
* Optional, Splunk Observability realm to which data will be sent to
* Default = us0

splunk_config = <value>
* Optional, config file that will be used by the Splunk_TA_OTel_Collector
* Default = $SPLUNK_HOME/etc/apps/Splunk_TA_OTel_Collector/configs/agent_config.yaml
