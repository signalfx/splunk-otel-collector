#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = r'''
---
module: otel_receiver_info
short_description: Query OpenTelemetry Collector receiver configurations
version_added: "1.0.0"
description:
  - Retrieves receiver configuration information from the collector config file.
  - Returns receiver metadata following Azure resource taxonomy.
  - Cross-references receivers with pipelines to show usage.
options:
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  name:
    description:
      - Filter results to only receivers matching this name.
    type: str
author:
  - Cisco OpenTelemetry Team
'''

EXAMPLES = r'''
- name: Get all receiver configurations
  signalfx.splunk_otel_collector.otel_receiver_info:
  register: result

- name: Get specific receiver
  signalfx.splunk_otel_collector.otel_receiver_info:
    name: otlp
  register: result

- name: Get receivers from custom config
  signalfx.splunk_otel_collector.otel_receiver_info:
    config_path: /etc/otel/config.yaml
  register: result

- name: Display receiver details
  ansible.builtin.debug:
    var: result.receivers
'''

RETURN = r'''
receivers:
  description: List of receiver configurations
  returned: always
  type: list
  elements: dict
  contains:
    id:
      description: Unique identifier (hostname:receiver_name)
      type: str
      sample: "web-server-01:otlp"
    name:
      description: Receiver name
      type: str
      sample: "otlp"
    location:
      description: Hostname where receiver is configured
      type: str
      sample: "web-server-01"
    tags:
      description: Metadata tags
      type: dict
      sample: {}
    properties:
      description: Receiver-specific properties
      type: dict
      contains:
        config:
          description: Receiver configuration dictionary
          type: dict
          sample: {"protocols": {"grpc": {"endpoint": "0.0.0.0:4317"}}}
        used_in_pipelines:
          description: List of pipeline names using this receiver
          type: list
          elements: str
          sample: ["traces", "metrics"]
    provisioning_state:
      description: Receiver configuration status
      type: str
      sample: "Configured"
'''

import socket

from ansible.module_utils.basic import AnsibleModule

try:
    from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.otel_config import OtelConfig
except ImportError:
    # Fallback for development/testing
    import sys
    import os
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'module_utils'))
    from otel_config import OtelConfig


def get_receiver_info(module):
    """Gather receiver information.

    Args:
        module: AnsibleModule instance

    Returns:
        dict: Result dictionary with receivers list
    """
    config_path = module.params['config_path']
    name_filter = module.params.get('name')

    hostname = socket.gethostname()

    # Load config
    config = OtelConfig(config_path)
    try:
        data = config.load()
    except Exception as e:
        module.fail_json(msg=f"Failed to load config: {str(e)}")

    receivers_config = data.get('receivers', {})
    pipelines_config = data.get('service', {}).get('pipelines', {})

    # Build reverse index: receiver -> pipelines
    receiver_usage = {}
    for pipeline_name, pipeline_config in pipelines_config.items():
        for receiver_name in pipeline_config.get('receivers', []):
            if receiver_name not in receiver_usage:
                receiver_usage[receiver_name] = []
            receiver_usage[receiver_name].append(pipeline_name)

    # Build receiver resources
    receivers = []
    for receiver_name, receiver_config in receivers_config.items():
        # Apply name filter
        if name_filter is not None and receiver_name != name_filter:
            continue

        receiver = {
            'id': f"{hostname}:{receiver_name}",
            'name': receiver_name,
            'location': hostname,
            'tags': {},
            'properties': {
                'config': receiver_config if receiver_config else {},
                'used_in_pipelines': receiver_usage.get(receiver_name, [])
            },
            'provisioning_state': 'Configured'
        }
        receivers.append(receiver)

    return {
        'changed': False,
        'receivers': receivers
    }


def main():
    """Entry point for module execution."""
    argument_spec = dict(
        config_path=dict(type='str', default='/etc/otel/collector/agent_config.yaml'),
        name=dict(type='str'),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    result = get_receiver_info(module)
    module.exit_json(**result)


if __name__ == '__main__':
    main()
