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
module: otel_collector_pipeline
short_description: Manage OpenTelemetry Collector pipelines
version_added: "1.0.0"
description:
  - Manages complete OpenTelemetry Collector pipelines including receivers, processors, and exporters.
  - Creates or updates the pipeline configuration in the collector's YAML config file.
  - Supports idempotent operations with change detection.
options:
  name:
    description:
      - Name of the pipeline (e.g., 'logs', 'traces', 'metrics').
    type: str
    required: true
  config_path:
    description:
      - Path to the OpenTelemetry Collector configuration file.
    type: str
    default: /etc/otel/collector/agent_config.yaml
  receivers:
    description:
      - List of receivers to configure for this pipeline.
    type: list
    elements: dict
    suboptions:
      name:
        description:
          - Name of the receiver.
        type: str
        required: true
      config:
        description:
          - Configuration dictionary for the receiver.
        type: dict
        default: {}
  processors:
    description:
      - List of processors to configure for this pipeline.
    type: list
    elements: dict
    suboptions:
      name:
        description:
          - Name of the processor.
        type: str
        required: true
      config:
        description:
          - Configuration dictionary for the processor.
        type: dict
        default: {}
  exporters:
    description:
      - List of exporters to configure for this pipeline.
    type: list
    elements: dict
    suboptions:
      name:
        description:
          - Name of the exporter.
        type: str
        required: true
      config:
        description:
          - Configuration dictionary for the exporter.
        type: dict
        default: {}
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
      - Desired state of the pipeline.
      - C(present) creates or updates the pipeline.
      - C(absent) removes the pipeline (leaves components since they may be used by other pipelines).
    type: str
    default: present
    choices: [ present, absent ]
author:
  - Cisco Systems
'''

EXAMPLES = r'''
- name: Create a logs pipeline with syslog receiver
  signalfx.splunk_otel_collector.otel_collector_pipeline:
    name: logs
    receivers:
      - name: syslog
        config:
          protocol: rfc5424
    processors:
      - name: batch
        config:
          timeout: 5s
      - name: memory_limiter
        config:
          limit_mib: 512
    exporters:
      - name: otlphttp
        config:
          endpoint: "https://gateway:4318"
    state: present

- name: Create a traces pipeline
  signalfx.splunk_otel_collector.otel_collector_pipeline:
    name: traces
    receivers:
      - name: otlp
        config:
          protocols:
            grpc:
              endpoint: "0.0.0.0:4317"
    processors:
      - name: batch
        config:
          timeout: 10s
    exporters:
      - name: otlphttp
        config:
          endpoint: "https://gateway:4318"
    backup: true
    state: present

- name: Remove a pipeline
  signalfx.splunk_otel_collector.otel_collector_pipeline:
    name: old_pipeline
    state: absent
'''

RETURN = r'''
changed:
  description: Whether the configuration was modified.
  type: bool
  returned: always
  sample: true
pipeline:
  description: The pipeline configuration.
  type: dict
  returned: when state is present
  sample:
    receivers: ["syslog"]
    processors: ["batch", "memory_limiter"]
    exporters: ["otlphttp"]
diff:
  description: Differences between the old and new configuration.
  type: dict
  returned: when changed
  sample:
    before: {}
    after:
      receivers: ["syslog"]
      processors: ["batch"]
      exporters: ["otlphttp"]
message:
  description: Human-readable message describing what was done.
  type: str
  returned: always
  sample: "Pipeline 'logs' created"
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
        receivers=dict(type='list', elements='dict', options=dict(
            name=dict(type='str', required=True),
            config=dict(type='dict', default={}),
        )),
        processors=dict(type='list', elements='dict', options=dict(
            name=dict(type='str', required=True),
            config=dict(type='dict', default={}),
        )),
        exporters=dict(type='list', elements='dict', options=dict(
            name=dict(type='str', required=True),
            config=dict(type='dict', default={}),
        )),
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

        pipeline_name = module.params['name']
        state = module.params['state']

        if state == 'present':
            # Get current pipeline state
            current_pipeline = config.get_pipeline(pipeline_name)

            # Prepare component lists
            receivers = module.params.get('receivers') or []
            processors = module.params.get('processors') or []
            exporters = module.params.get('exporters') or []

            # Set each component
            for receiver in receivers:
                config.set_component('receivers', receiver['name'], receiver.get('config', {}))

            for processor in processors:
                config.set_component('processors', processor['name'], processor.get('config', {}))

            for exporter in exporters:
                config.set_component('exporters', exporter['name'], exporter.get('config', {}))

            # Build pipeline configuration
            receiver_names = [r['name'] for r in receivers]
            processor_names = [p['name'] for p in processors]
            exporter_names = [e['name'] for e in exporters]

            new_pipeline = {
                'receivers': receiver_names,
                'processors': processor_names,
                'exporters': exporter_names
            }

            # Check if pipeline changed
            changed = current_pipeline != new_pipeline

            if changed and not module.check_mode:
                config.set_pipeline(pipeline_name, receiver_names, processor_names, exporter_names)
                config.save(backup=module.params['backup'])

            result['changed'] = changed
            result['pipeline'] = new_pipeline

            if changed:
                result['diff'] = {
                    'before': current_pipeline or {},
                    'after': new_pipeline
                }
                result['message'] = f"Pipeline '{pipeline_name}' {'updated' if current_pipeline else 'created'}"
            else:
                result['message'] = f"Pipeline '{pipeline_name}' already configured"

        else:  # state == 'absent'
            current_pipeline = config.get_pipeline(pipeline_name)

            if current_pipeline:
                result['changed'] = True
                if not module.check_mode:
                    config.remove_pipeline(pipeline_name)
                    config.save(backup=module.params['backup'])

                result['diff'] = {
                    'before': current_pipeline,
                    'after': {}
                }
                result['message'] = f"Pipeline '{pipeline_name}' removed"
            else:
                result['message'] = f"Pipeline '{pipeline_name}' already absent"

        module.exit_json(**result)

    except Exception as e:
        module.fail_json(msg=f"Error managing pipeline: {str(e)}", **result)


def main():
    """Entry point."""
    run_module()


if __name__ == '__main__':
    main()
