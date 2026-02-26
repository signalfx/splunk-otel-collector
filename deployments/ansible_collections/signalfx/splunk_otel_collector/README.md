# Splunk OpenTelemetry Collector Ansible Collection

## Description

Ansible Collection `signalfx.splunk_otel_collector` contains just one [Ansible 
role for Splunk OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible_collections/signalfx/splunk_otel_collector/roles/collector): 
`signalfx.splunk_otel_collector.collector`.

The role installs Splunk OpenTelemetry Collector configured to
collect metrics, traces and logs from Linux and Windows machines and send data to [Splunk
Observability Cloud](https://www.splunk.com/en_us/products/observability.html). 

## Requirements

### Dependencies

- Python 3.11+

### Linux

Currently, the following Linux distributions and versions are supported:

- Amazon Linux: 2023
- CentOS / Red Hat: 9
- Debian: 11, 12
- Ubuntu: 22.04, 24.04

### Windows

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

## Use Cases

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
        collector_splunk_access_token: YOUR_ACCESS_TOKEN
        collector_splunk_hec_token: YOUR_HEC_TOKEN
        collector_splunk_realm: SPLUNK_REALM
```

> **_NOTE:_**  Setting collector_splunk_hec_token is optional.

Full documentation on how to configure the role:
[Splunk OpenTelemetry Collector Role](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible_collections/signalfx/splunk_otel_collector/roles/collector)

## Testing

Molecule is used to test the role with many different configuration options, including custom variables, various
combinations of auto instrumentation, and installing from a remote vs. local binary. These tests are run on
varying flavours of Linux and Windows to ensure compatibility.

## Contributing

Check [Contributing guidelines](https://github.com/signalfx/splunk-otel-collector/tree/main/deployments/ansible_collections/signalfx/splunk_otel_collector/contributing/README.md) 
if you see something that needs to be improved in this Ansible role.

## Support

- [Splunk Ideas](https://ideas.splunk.com/) - To request any enhancement, component addition, or new feature,
  please submit a new Splunk Idea. Requires a Splunk.com login.
- [Splunk Support](https://splunk.my.site.com/customer/s/need-help/create-case) - Ask any question or report any issue
  by opening a support case. Requires Splunk Support entitlement.
- As Red Hat Ansible Certified Content, this collection is entitled to support through Ansible Automation Platform (AAP). Use the Create issue button on the top right corner of the Automation Hub Collection page for any defects, feature requests, or questions on usage.

## Release Notes

[Release notes](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/ansible_collections/signalfx/splunk_otel_collector/CHANGELOG.md)

## Related Information

- [Splunk distribution of the OpenTelemetry Collector](https://github.com/signalfx/splunk-otel-collector)
- [Deploy the Collector for Linux With Ansible](https://help.splunk.com/en/splunk-observability-cloud/manage-data/splunk-distribution-of-the-opentelemetry-collector/get-started-with-the-splunk-distribution-of-the-opentelemetry-collector/collector-for-linux/install-the-collector-for-linux-tools/ansible-for-linux)
- [Using Ansible collections](https://docs.ansible.com/projects/ansible/latest/collections_guide/index.html)

## License Information

[Apache Software License version 2.0](https://github.com/signalfx/splunk-otel-collector/tree/main/LICENSE).
