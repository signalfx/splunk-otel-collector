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
module: otel_collector_info
short_description: Query OpenTelemetry Collector status information
version_added: "1.0.0"
description:
  - Retrieves status information about the OpenTelemetry Collector instance.
  - Queries the collector's health check endpoint and reads configuration.
  - Returns collector metadata following Azure resource taxonomy.
options:
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  health_endpoint:
    description:
      - URL of the collector's health check endpoint.
    type: str
    default: http://localhost:13133
  name:
    description:
      - Filter results to only collectors matching this name.
    type: str
author:
  - Cisco OpenTelemetry Team
'''

EXAMPLES = r'''
- name: Get collector status
  signalfx.splunk_otel_collector.otel_collector_info:
  register: result

- name: Get collector with custom health endpoint
  signalfx.splunk_otel_collector.otel_collector_info:
    health_endpoint: http://localhost:13133
    config_path: /etc/otel/config.yaml
  register: result

- name: Display collector status
  ansible.builtin.debug:
    var: result.collectors
'''

RETURN = r'''
collectors:
  description: List of collector information
  returned: always
  type: list
  elements: dict
  contains:
    id:
      description: Unique identifier (machine-id)
      type: str
      sample: "a1b2c3d4e5f6"
    name:
      description: Collector name
      type: str
      sample: "splunk-otel-collector"
    location:
      description: Hostname where collector is running
      type: str
      sample: "web-server-01"
    tags:
      description: Metadata tags
      type: dict
      sample: {}
    properties:
      description: Collector-specific properties
      type: dict
      contains:
        health_endpoint:
          description: Health check endpoint URL
          type: str
          sample: "http://localhost:13133"
        config_path:
          description: Path to configuration file
          type: str
          sample: "/etc/otel/collector/agent_config.yaml"
        active_pipelines:
          description: List of active pipeline names
          type: list
          elements: str
          sample: ["logs", "metrics", "traces"]
    provisioning_state:
      description: Collector operational status
      type: str
      sample: "Running"
      choices:
        - Running
        - Stopped
        - Unknown
'''

import socket
from pathlib import Path

try:
    import urllib.request
    import urllib.error
    HAS_URLLIB = True
except ImportError:
    HAS_URLLIB = False

from ansible.module_utils.basic import AnsibleModule

try:
    from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.otel_config import OtelConfig
except ImportError:
    # Fallback for development/testing
    import sys
    import os
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'module_utils'))
    from otel_config import OtelConfig


def get_machine_id():
    """Read machine-id from standard Linux location.

    Returns:
        str: Machine ID or "unknown" if not available
    """
    machine_id_paths = [
        '/etc/machine-id',
        '/var/lib/dbus/machine-id'
    ]

    for path in machine_id_paths:
        try:
            with open(path, 'r', encoding='utf-8') as f:
                return f.read().strip()
        except (IOError, OSError):
            continue

    return "unknown"


def check_health(endpoint):
    """Query the collector health endpoint.

    Args:
        endpoint: Health endpoint URL

    Returns:
        str: "Running" if healthy, "Stopped" if connection refused, "Unknown" otherwise
    """
    if not HAS_URLLIB:
        return "Unknown"

    try:
        with urllib.request.urlopen(endpoint, timeout=5) as response:
            if response.status == 200:
                return "Running"
            return "Unknown"
    except urllib.error.URLError as e:
        # Connection refused means service is stopped
        if hasattr(e, 'reason') and 'Connection refused' in str(e.reason):
            return "Stopped"
        return "Unknown"
    except Exception:
        return "Unknown"


def get_collector_info(module):
    """Gather collector information.

    Args:
        module: AnsibleModule instance

    Returns:
        dict: Result dictionary with collectors list
    """
    config_path = module.params['config_path']
    health_endpoint = module.params['health_endpoint']
    name_filter = module.params.get('name')

    # Get system information
    machine_id = get_machine_id()
    hostname = socket.gethostname()

    # Load config to determine active pipelines
    config = OtelConfig(config_path)
    try:
        data = config.load()
        active_pipelines = list(data.get('service', {}).get('pipelines', {}).keys())
    except Exception as e:
        module.fail_json(msg=f"Failed to load config: {str(e)}")

    # Check health status
    provisioning_state = check_health(health_endpoint)

    # Build collector resource
    collector = {
        'id': machine_id,
        'name': 'splunk-otel-collector',
        'location': hostname,
        'tags': {},
        'properties': {
            'health_endpoint': health_endpoint,
            'config_path': config_path,
            'active_pipelines': active_pipelines
        },
        'provisioning_state': provisioning_state
    }

    # Apply name filter
    collectors = []
    if name_filter is None or collector['name'] == name_filter:
        collectors.append(collector)

    return {
        'changed': False,
        'collectors': collectors
    }


def main():
    """Entry point for module execution."""
    argument_spec = dict(
        config_path=dict(type='str', default='/etc/otel/collector/agent_config.yaml'),
        health_endpoint=dict(type='str', default='http://localhost:13133'),
        name=dict(type='str'),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
    )

    result = get_collector_info(module)
    module.exit_json(**result)


if __name__ == '__main__':
    main()
