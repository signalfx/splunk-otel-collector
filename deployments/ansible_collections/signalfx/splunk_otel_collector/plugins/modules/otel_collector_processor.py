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
module: otel_collector_processor
short_description: Manage OpenTelemetry Collector processors
version_added: "1.0.0"
description:
  - Manages individual OpenTelemetry Collector processors.
  - Creates, updates, or removes processor configurations in the collector's YAML config file.
  - Optionally adds the processor to specified pipelines.
  - Supports idempotent operations with change detection.
options:
  name:
    description:
      - Name of the processor (e.g., 'batch', 'memory_limiter', 'attributes').
    type: str
    required: true
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  config:
    description:
      - Configuration dictionary for the processor.
    type: dict
    default: {}
  pipelines:
    description:
      - List of pipeline names to add this processor to.
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
      - Desired state of the processor.
      - C(present) creates or updates the processor.
      - C(absent) removes the processor and its pipeline references.
    type: str
    default: present
    choices: [ present, absent ]
author:
  - Cisco Systems
'''

EXAMPLES = r'''
- name: Create a batch processor
  signalfx.splunk_otel_collector.otel_collector_processor:
    name: batch
    config:
      timeout: 10s
      send_batch_size: 1024
    pipelines:
      - traces
      - metrics
      - logs
    state: present

- name: Create a memory limiter processor
  signalfx.splunk_otel_collector.otel_collector_processor:
    name: memory_limiter
    config:
      limit_mib: 512
      spike_limit_mib: 128
      check_interval: 5s
    pipelines:
      - logs
    state: present

- name: Update a processor configuration
  signalfx.splunk_otel_collector.otel_collector_processor:
    name: attributes
    config:
      actions:
        - key: environment
          value: production
          action: upsert
    state: present

- name: Remove a processor
  signalfx.splunk_otel_collector.otel_collector_processor:
    name: old_processor
    state: absent
'''

RETURN = r'''
changed:
  description: Whether the configuration was modified.
  type: bool
  returned: always
  sample: true
processor:
  description: The processor configuration.
  type: dict
  returned: when state is present
  sample:
    timeout: 10s
    send_batch_size: 1024
diff:
  description: Differences between the old and new configuration.
  type: dict
  returned: when changed
  sample:
    before: {}
    after:
      timeout: 10s
      send_batch_size: 1024
message:
  description: Human-readable message describing what was done.
  type: str
  returned: always
  sample: "Processor 'batch' created and added to pipelines: traces, metrics"
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

        processor_name = module.params['name']
        state = module.params['state']

        if state == 'present':
            # Get current processor state
            current_processor = config.get_component('processors', processor_name)
            new_processor = module.params['config']

            # Check if processor config changed
            processor_changed = current_processor != new_processor

            # Update pipelines
            pipelines = module.params.get('pipelines') or []
            pipeline_changes = []

            for pipeline_name in pipelines:
                pipeline = config.get_pipeline(pipeline_name)
                if pipeline:
                    if processor_name not in pipeline.get('processors', []):
                        pipeline_changes.append(pipeline_name)

            changed = processor_changed or bool(pipeline_changes)

            if changed and not module.check_mode:
                # Set the processor
                config.set_component('processors', processor_name, new_processor)

                # Add to pipelines
                for pipeline_name in pipeline_changes:
                    pipeline = config.get_pipeline(pipeline_name)
                    if pipeline:
                        processors_list = list(pipeline.get('processors', []))
                        processors_list.append(processor_name)
                        config.set_pipeline(
                            pipeline_name,
                            pipeline.get('receivers', []),
                            processors_list,
                            pipeline.get('exporters', [])
                        )

                config.save(backup=module.params['backup'])

            result['changed'] = changed
            result['processor'] = new_processor

            if changed:
                result['diff'] = {
                    'before': current_processor or {},
                    'after': new_processor
                }
                msg_parts = [f"Processor '{processor_name}' {'updated' if current_processor else 'created'}"]
                if pipeline_changes:
                    msg_parts.append(f"added to pipelines: {', '.join(pipeline_changes)}")
                result['message'] = ' and '.join(msg_parts)
            else:
                result['message'] = f"Processor '{processor_name}' already configured"

        else:  # state == 'absent'
            current_processor = config.get_component('processors', processor_name)

            if current_processor:
                result['changed'] = True
                if not module.check_mode:
                    config.remove_component('processors', processor_name)
                    config.save(backup=module.params['backup'])

                result['diff'] = {
                    'before': current_processor,
                    'after': {}
                }
                result['message'] = f"Processor '{processor_name}' removed"
            else:
                result['message'] = f"Processor '{processor_name}' already absent"

        module.exit_json(**result)

    except Exception as e:
        module.fail_json(msg=f"Error managing processor: {str(e)}", **result)


def main():
    """Entry point."""
    run_module()


if __name__ == '__main__':
    main()
