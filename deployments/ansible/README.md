# Ansible Collection for Splunk OpenTelemetry Collector

Ansible Collection `signalfx.splunk_otel_collector` contains just one [Ansible 
role for Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/roles/collector): 
`signalfx.splunk_otel_collector.collector`.

The role installs Splunk OpenTelemetry Collector configured to
collect metrics, traces and logs from Linux machines and send data to [Splunk 
Observability Cloud](https://www.splunk.com/en_us/products/observability.html). 

## Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2, 2023
- CentOS / Red Hat: 8, 9
- Oracle: 8, 9
- Debian: 9, 10, 11
- SUSE: 15
- Ubuntu: 16.04, 18.04, 20.04, 22.04

## Windows

Currently, the following Windows versions are supported:

- Windows Server 2016 64-bit
- Windows Server 2019 64-bit
- Windows Server 2022 64-bit

On Windows, the collector is installed as a Windows service and its environment
variables are set at the service scope, i.e.: they are only available to the
collector service and not to the entire machine.

Ansible requires PowerShell 3.0 or newer and at least .NET 4.0 to be installed on Windows host.
A WinRM listener should be created and activated.
For setting up Windows Host refer [Ansible Docs](https://docs.ansible.com/ansible/latest/user_guide/windows_setup.html).

## Installation

To install the Ansible collection from Ansible Galaxy:

```sh
ansible-galaxy collection install signalfx.splunk_otel_collector
```

## Usage

To use Splunk OpenTelemetry Collector Role, simply include the 
`signalfx.splunk_otel_collector.collector` role invocation in your playbook. 
Note that this role requires root access.

```yaml
- name: Install Splunk OpenTelemetry Collector
  hosts: all
  become: yes
  # For Windows "become: yes" will raise error.
  # "The Powershell family is incompatible with the sudo become plugin". Remove "become: yes" tag to run on Windows
  tasks:
    - name: "Include splunk_otel_collector"
      ansible.builtin.include_role:
        name: "signalfx.splunk_otel_collector.collector"
      vars:
        splunk_access_token: YOUR_ACCESS_TOKEN
        splunk_hec_token: YOUR_HEC_TOKEN
        splunk_realm: SPLUNK_REALM
```

> **_NOTE:_**  Setting splunk_hec_token is optional.

Full documentation on how to configure the role:
[Splunk OpenTelemetry Collector Role](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/roles/collector)

## Contributing

Check [Contributing guidelines](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible/contributing/README.md) 
if you see something that needs to be improved in this Ansible role.

## License

[Apache Software License version 2.0](https://github.com/signalfx/splunk-otel-collector/tree/main/LICENSE).
