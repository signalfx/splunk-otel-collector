# Ansible Collection for Splunk OpenTelemetry Connector

Ansible Collection `signalfx.splunk_otel_collector` contains just one [Ansible 
role for Splunk OpenTelemetry Connector](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/roles/collector): 
`signalfx.splunk_otel_collector.collector`.

The role installs Splunk OpenTelemetry Connector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/observability.html). 

## Linux
Currently, the following Linux distributions and versions are supported:

- Amazon Linux
- Debian: buster, stretch
- Ubuntu: 16.04, 18.04, 20.04

## Windows
Currently, the following Windows versions are supported:
Ansible requires PowerShell 3.0 or newer and atleast .NET4.0 to be installed on Windows host.
A WinRM listner should be created and activeted. 
For setting up Windows Host Refer:[Ansible Docs](https://docs.ansible.com/ansible/latest/user_guide/windows_setup.html).

- Windows Server 2012 64-bit
- Windows Server 2016 64-bit
- Windows Server 2019 64-bit

## Installation

To install the Ansible collection from Ansible Galaxy:
```sh
ansible-galaxy collection install signalfx.splunk_otel_collector
```

## Usage

To use Splunk OpenTelemetry Connector Role, simply include the 
`signalfx.splunk_otel_collector.collector` role invocation in your playbook. 
Note that this role requires root access.

```yaml
- name: Install Splunk OpenTelemetry Connector
  hosts: all
  # For Windows "become: yes" will raise error.
  # "The Powershell family is incompatible with the sudo become plugin" Remove "become: yes" tag to run on Windows
  become: yes
  tasks:
    - name: "Include splunk_otel_collector"
      include_role:
        name: "signalfx.splunk_otel_collector.collector"
      vars:
        splunk_access_token: YOUR_ACCESS_TOKEN
        splunk_realm: SPLUNK_REALM
```

Full documentation on how to configure the role:
[Splunk OpenTelemetry Connector Role](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/roles/collector)

## Contributing

Check [Contributing guidelines](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/contributing/README.md) 
if you see something that needs to be improved in this Ansible role.

## License

[Apache Software License version 2.0](https://github.com/signalfx/splunk-otel-collector/tree/main/LICENSE).
