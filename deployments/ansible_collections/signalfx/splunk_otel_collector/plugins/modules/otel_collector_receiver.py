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
module: otel_collector_receiver
short_description: Manage OpenTelemetry Collector receivers
version_added: "1.0.0"
description:
  - Manages individual OpenTelemetry Collector receivers.
  - Creates, updates, or removes receiver configurations in the collector's YAML config file.
  - Optionally adds the receiver to specified pipelines.
  - Supports idempotent operations with change detection.
options:
  name:
    description:
      - Name of the receiver (e.g., 'otlp', 'syslog', 'prometheus').
    type: str
    required: true
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  config:
    description:
      - Configuration dictionary for the receiver.
    type: dict
    default: {}
  pipelines:
    description:
      - List of pipeline names to add this receiver to.
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
      - Desired state of the receiver.
      - C(present) creates or updates the receiver.
      - C(absent) removes the receiver and its pipeline references.
    type: str
    default: present
    choices: [ present, absent ]
author:
  - Cisco Systems
'''

EXAMPLES = r'''
- name: Create an OTLP receiver
  signalfx.splunk_otel_collector.otel_collector_receiver:
    name: otlp
    config:
      protocols:
        grpc:
          endpoint: "0.0.0.0:4317"
        http:
          endpoint: "0.0.0.0:4318"
    pipelines:
      - traces
      - metrics
    state: present

- name: Create a syslog receiver
  signalfx.splunk_otel_collector.otel_collector_receiver:
    name: syslog
    config:
      protocol: rfc5424
      location: 0.0.0.0:514
    pipelines:
      - logs
    state: present

- name: Update a receiver configuration
  signalfx.splunk_otel_collector.otel_collector_receiver:
    name: prometheus
    config:
      endpoint: "0.0.0.0:9090"
      scrape_interval: 30s
    state: present

- name: Remove a receiver
  signalfx.splunk_otel_collector.otel_collector_receiver:
    name: old_receiver
    state: absent
'''

RETURN = r'''
changed:
  description: Whether the configuration was modified.
  type: bool
  returned: always
  sample: true
receiver:
  description: The receiver configuration.
  type: dict
  returned: when state is present
  sample:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
diff:
  description: Differences between the old and new configuration.
  type: dict
  returned: when changed
  sample:
    before: {}
    after:
      protocols:
        grpc:
          endpoint: "0.0.0.0:4317"
message:
  description: Human-readable message describing what was done.
  type: str
  returned: always
  sample: "Receiver 'otlp' created and added to pipelines: traces, metrics"
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

        receiver_name = module.params['name']
        state = module.params['state']

        if state == 'present':
            # Get current receiver state
            current_receiver = config.get_component('receivers', receiver_name)
            new_receiver = module.params['config']

            # Check if receiver config changed
            receiver_changed = current_receiver != new_receiver

            # Update pipelines
            pipelines = module.params.get('pipelines') or []
            pipeline_changes = []

            for pipeline_name in pipelines:
                pipeline = config.get_pipeline(pipeline_name)
                if pipeline:
                    if receiver_name not in pipeline.get('receivers', []):
                        pipeline_changes.append(pipeline_name)

            changed = receiver_changed or bool(pipeline_changes)

            if changed and not module.check_mode:
                # Set the receiver
                config.set_component('receivers', receiver_name, new_receiver)

                # Add to pipelines
                for pipeline_name in pipeline_changes:
                    pipeline = config.get_pipeline(pipeline_name)
                    if pipeline:
                        receivers_list = list(pipeline.get('receivers', []))
                        receivers_list.append(receiver_name)
                        config.set_pipeline(
                            pipeline_name,
                            receivers_list,
                            pipeline.get('processors', []),
                            pipeline.get('exporters', [])
                        )

                config.save(backup=module.params['backup'])

            result['changed'] = changed
            result['receiver'] = new_receiver

            if changed:
                result['diff'] = {
                    'before': current_receiver or {},
                    'after': new_receiver
                }
                msg_parts = [f"Receiver '{receiver_name}' {'updated' if current_receiver else 'created'}"]
                if pipeline_changes:
                    msg_parts.append(f"added to pipelines: {', '.join(pipeline_changes)}")
                result['message'] = ' and '.join(msg_parts)
            else:
                result['message'] = f"Receiver '{receiver_name}' already configured"

        else:  # state == 'absent'
            current_receiver = config.get_component('receivers', receiver_name)

            if current_receiver:
                result['changed'] = True
                if not module.check_mode:
                    config.remove_component('receivers', receiver_name)
                    config.save(backup=module.params['backup'])

                result['diff'] = {
                    'before': current_receiver,
                    'after': {}
                }
                result['message'] = f"Receiver '{receiver_name}' removed"
            else:
                result['message'] = f"Receiver '{receiver_name}' already absent"

        module.exit_json(**result)

    except Exception as e:
        module.fail_json(msg=f"Error managing receiver: {str(e)}", **result)


def main():
    """Entry point."""
    run_module()


if __name__ == '__main__':
    main()
