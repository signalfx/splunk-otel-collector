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
module: otel_collector_exporter
short_description: Manage OpenTelemetry Collector exporters
version_added: "1.0.0"
description:
  - Manages individual OpenTelemetry Collector exporters.
  - Creates, updates, or removes exporter configurations in the collector's YAML config file.
  - Optionally adds the exporter to specified pipelines.
  - Supports idempotent operations with change detection.
options:
  name:
    description:
      - Name of the exporter (e.g., 'otlphttp', 'logging', 'prometheus').
    type: str
    required: true
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  config:
    description:
      - Configuration dictionary for the exporter.
    type: dict
    default: {}
  pipelines:
    description:
      - List of pipeline names to add this exporter to.
      - Only applies when state is present.
    type: list
    elements: str
  restart_collector:
    description:
      - Whether to restart the collector service after configuration changes.
    type: bool
    default: false
  backup:
    description:
      - Whether to create a timestamped backup of the config file before modifying.
    type: bool
    default: true
  state:
    description:
      - Desired state of the exporter.
      - C(present) creates or updates the exporter.
      - C(absent) removes the exporter and its pipeline references.
    type: str
    default: present
    choices: [ present, absent ]
author:
  - Cisco Systems
'''

EXAMPLES = r'''
- name: Create an OTLP HTTP exporter
  signalfx.splunk_otel_collector.otel_collector_exporter:
    name: otlphttp
    config:
      endpoint: "https://gateway:4318"
      headers:
        api-key: "secret-key"
    pipelines:
      - traces
      - metrics
      - logs
    state: present

- name: Create a logging exporter
  signalfx.splunk_otel_collector.otel_collector_exporter:
    name: logging
    config:
      loglevel: debug
    pipelines:
      - traces
    state: present

- name: Update an exporter configuration
  signalfx.splunk_otel_collector.otel_collector_exporter:
    name: prometheus
    config:
      endpoint: "0.0.0.0:8889"
      namespace: "otel"
    state: present

- name: Remove an exporter
  signalfx.splunk_otel_collector.otel_collector_exporter:
    name: old_exporter
    state: absent
'''

RETURN = r'''
changed:
  description: Whether the configuration was modified.
  type: bool
  returned: always
  sample: true
exporter:
  description: The exporter configuration.
  type: dict
  returned: when state is present
  sample:
    endpoint: "https://gateway:4318"
    headers:
      api-key: "secret-key"
diff:
  description: Differences between the old and new configuration.
  type: dict
  returned: when changed
  sample:
    before: {}
    after:
      endpoint: "https://gateway:4318"
message:
  description: Human-readable message describing what was done.
  type: str
  returned: always
  sample: "Exporter 'otlphttp' created and added to pipelines: traces, metrics"
'''

from ansible.module_utils.basic import AnsibleModule

try:
    from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.otel_config import OtelConfig
except ImportError:
    # Fallback for development/testing
    import sys
    import os
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'module_utils'))
    from otel_config import OtelConfig


def run_module():
    """Main module execution."""
    module_args = dict(
        name=dict(type='str', required=True),
        config_path=dict(type='str', default='/etc/otel/collector/agent_config.yaml'),
        config=dict(type='dict', default={}),
        pipelines=dict(type='list', elements='str'),
        restart_collector=dict(type='bool', default=False),
        backup=dict(type='bool', default=True),
        state=dict(type='str', default='present', choices=['present', 'absent']),
    )

    result = dict(
        changed=False,
        message='',
    )

    module = AnsibleModule(
        argument_spec=module_args,
        supports_check_mode=True
    )

    try:
        config = OtelConfig(module.params['config_path'])
        config.load()

        exporter_name = module.params['name']
        state = module.params['state']

        if state == 'present':
            # Get current exporter state
            current_exporter = config.get_component('exporters', exporter_name)
            new_exporter = module.params['config']

            # Check if exporter config changed
            exporter_changed = current_exporter != new_exporter

            # Update pipelines
            pipelines = module.params.get('pipelines') or []
            pipeline_changes = []

            for pipeline_name in pipelines:
                pipeline = config.get_pipeline(pipeline_name)
                if pipeline:
                    if exporter_name not in pipeline.get('exporters', []):
                        pipeline_changes.append(pipeline_name)

            changed = exporter_changed or bool(pipeline_changes)

            if changed and not module.check_mode:
                # Set the exporter
                config.set_component('exporters', exporter_name, new_exporter)

                # Add to pipelines
                for pipeline_name in pipeline_changes:
                    pipeline = config.get_pipeline(pipeline_name)
                    if pipeline:
                        exporters_list = list(pipeline.get('exporters', []))
                        exporters_list.append(exporter_name)
                        config.set_pipeline(
                            pipeline_name,
                            pipeline.get('receivers', []),
                            pipeline.get('processors', []),
                            exporters_list
                        )

                config.save(backup=module.params['backup'])

            result['changed'] = changed
            result['exporter'] = new_exporter

            if changed:
                result['diff'] = {
                    'before': current_exporter or {},
                    'after': new_exporter
                }
                msg_parts = [f"Exporter '{exporter_name}' {'updated' if current_exporter else 'created'}"]
                if pipeline_changes:
                    msg_parts.append(f"added to pipelines: {', '.join(pipeline_changes)}")
                result['message'] = ' and '.join(msg_parts)
            else:
                result['message'] = f"Exporter '{exporter_name}' already configured"

        else:  # state == 'absent'
            current_exporter = config.get_component('exporters', exporter_name)

            if current_exporter:
                result['changed'] = True
                if not module.check_mode:
                    config.remove_component('exporters', exporter_name)
                    config.save(backup=module.params['backup'])

                result['diff'] = {
                    'before': current_exporter,
                    'after': {}
                }
                result['message'] = f"Exporter '{exporter_name}' removed"
            else:
                result['message'] = f"Exporter '{exporter_name}' already absent"

        module.exit_json(**result)

    except Exception as e:
        module.fail_json(msg=f"Error managing exporter: {str(e)}", **result)


def main():
    """Entry point."""
    run_module()


if __name__ == '__main__':
    main()
