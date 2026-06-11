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
module: otel_pipeline_info
short_description: Query OpenTelemetry Collector pipeline configurations
version_added: "1.0.0"
description:
  - Retrieves pipeline configuration information from the collector config file.
  - Returns pipeline metadata following Azure resource taxonomy.
  - Parses service.pipelines section to extract receivers, processors, and exporters.
options:
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  name:
    description:
      - Filter results to only pipelines matching this name.
    type: str
author:
  - Cisco OpenTelemetry Team
'''

EXAMPLES = r'''
- name: Get all pipeline configurations
  signalfx.splunk_otel_collector.otel_pipeline_info:
  register: result

- name: Get specific pipeline
  signalfx.splunk_otel_collector.otel_pipeline_info:
    name: logs
  register: result

- name: Get pipelines from custom config
  signalfx.splunk_otel_collector.otel_pipeline_info:
    config_path: /etc/otel/config.yaml
  register: result

- name: Display pipeline details
  ansible.builtin.debug:
    var: result.pipelines
'''

RETURN = r'''
pipelines:
  description: List of pipeline configurations
  returned: always
  type: list
  elements: dict
  contains:
    id:
      description: Unique identifier (hostname:pipeline_name)
      type: str
      sample: "web-server-01:logs"
    name:
      description: Pipeline name
      type: str
      sample: "logs"
    location:
      description: Hostname where pipeline is configured
      type: str
      sample: "web-server-01"
    tags:
      description: Metadata tags
      type: dict
      sample: {}
    properties:
      description: Pipeline-specific properties
      type: dict
      contains:
        signal_type:
          description: Type of telemetry signal
          type: str
          sample: "logs"
          choices:
            - logs
            - metrics
            - traces
            - unknown
        receivers:
          description: List of receiver names in this pipeline
          type: list
          elements: str
          sample: ["syslog", "filelog"]
        processors:
          description: List of processor names in this pipeline
          type: list
          elements: str
          sample: ["batch", "memory_limiter"]
        exporters:
          description: List of exporter names in this pipeline
          type: list
          elements: str
          sample: ["otlphttp"]
    provisioning_state:
      description: Pipeline configuration status
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


def infer_signal_type(pipeline_name):
    """Infer signal type from pipeline name.

    Args:
        pipeline_name: Name of the pipeline

    Returns:
        str: One of 'logs', 'metrics', 'traces', or 'unknown'
    """
    name_lower = pipeline_name.lower()
    if 'log' in name_lower:
        return 'logs'
    elif 'metric' in name_lower:
        return 'metrics'
    elif 'trace' in name_lower or 'span' in name_lower:
        return 'traces'
    else:
        return 'unknown'


def get_pipeline_info(module):
    """Gather pipeline information.

    Args:
        module: AnsibleModule instance

    Returns:
        dict: Result dictionary with pipelines list
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

    pipelines_config = data.get('service', {}).get('pipelines', {})

    # Build pipeline resources
    pipelines = []
    for pipeline_name, pipeline_config in pipelines_config.items():
        # Apply name filter
        if name_filter is not None and pipeline_name != name_filter:
            continue

        pipeline = {
            'id': f"{hostname}:{pipeline_name}",
            'name': pipeline_name,
            'location': hostname,
            'tags': {},
            'properties': {
                'signal_type': infer_signal_type(pipeline_name),
                'receivers': pipeline_config.get('receivers', []),
                'processors': pipeline_config.get('processors', []),
                'exporters': pipeline_config.get('exporters', [])
            },
            'provisioning_state': 'Configured'
        }
        pipelines.append(pipeline)

    return {
        'changed': False,
        'pipelines': pipelines
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

    result = get_pipeline_info(module)
    module.exit_json(**result)


if __name__ == '__main__':
    main()
