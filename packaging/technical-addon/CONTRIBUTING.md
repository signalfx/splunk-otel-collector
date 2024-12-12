# Orca
(For internal splunkers only)
To use orca,
0. Set up a venv `mkdir -p ~/.venvs; python3 -m venv ~/.venvs/orca`
1. Activate the venv `source ~/.venvs/orca/bin/activate`
2. Ensure you're on vpn and have authenticated with artifactory (ex `okta-artifactory-login`)
3. Install orca `pip install --upgrade orca`

# Testing
Currently, we use orca, a wrapper for splunk-ansible.

To auth into splunk orca, run `splunk_orca config auth`.  It's usually best to install orca to a virtual env (assuming you've run `okta-artifactory-login`)
```
python3.9 -m venv ../.orca-latest && source ../.orca-latest/bin/activate && pip install --upgrade splunk_orca
```

To test locally, grab an access token and run something like

```
make install-tools && PLATFORM=all make distribute-ta && OLLY_ACCESS_TOKEN="<REDACTED>" UF_VERSION=9.0.2 SPLUNK_PLATFORM=x64_windows_2022 PLATFORM=windows ARCH=amd64 ORCA_OPTION="" ORCA_CLOUD="aws" SPLUNK_CONFIG='$SPLUNK_OTEL_TA_HOME/configs/ta-agent-config.yaml' BUILD_DIR=$(pwd)/build make orca-test-ta
```

If you want to test windows, you may need to run something like 
```
 rm -rf build && PLATFORM=all make distribute-ta && OLLY_ACCESS_TOKEN="<REDACTED>" UF_VERSION=9.0.2 SPLUNK_PLATFORM=x64_windows_2022 PLATFORM=windows ARCH=amd64 ORCA_OPTION="" ORCA_CLOUD="aws" make -e orca-test-ta
```

When debugging orca itself, you can directly invoke provide a local path to your TA and any ansible as such:
```
splunk_orca -vvv --cloud aws --printer sdd-json --deployment-file /home/jameshughes/workspace/otel-github/splunk-otel-collector/build/orca_deployment.json --ansible-log ansible-local.log create --prefix happypath --env SPLUNK_CONNECTION_TIMEOUT=600 --splunk-version 9.0.2 --platform x64_windows_2022 --local-apps /home/jameshughes/workspace/otel-github/splunk-otel-collector/build/ci-cd/Splunk_TA_otel.tgz
```

