[Splunk_TA_OTel_Collector://<name>]

splunk_access_token = <value>
* Access token used to send data to Splunk Observability
* Default =

splunk_realm = <value>
* Splunk Observability realm to which data will be sent to
* Default =

splunk_config = <value>
* Config file that will be used by the Splunk_TA_OTel_Collector
* Default = $SPLUNK_HOME/etc/apps/Splunk_TA_OTel_Collector/configs/agent_config.yaml

splunk_collector_env_vars = <value>
* Specifies the environment variables for the Splunk Collector. The value should be in the format of "KEY1=VALUE1,KEY2=VALUE2". The ',' and '=' characters in values MUST be percent encoded.
* Default =

splunk_collector_cmd_args = <value>
* Specifies the command arguments for the Splunk Collector
* Default =

