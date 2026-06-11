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
module: otel_node_count
short_description: Count unique managed nodes across multiple discovery sources
version_added: "1.0.0"
description:
  - Queries multiple discovery sources to determine the unique count of managed nodes.
  - Supports OpAMP servers, collector health APIs, Ansible inventory, and Kubernetes.
  - Deduplicates nodes by host.id (machine-id) to provide accurate indirect node licensing counts.
  - Returns node details in Azure Resource Manager taxonomy format.
author:
  - Cisco Systems, Inc.
options:
  sources:
    description:
      - List of discovery source configurations.
      - Each source must specify a type and type-specific parameters.
    type: list
    elements: dict
    required: true
    suboptions:
      type:
        description:
          - Type of discovery source.
        type: str
        required: true
        choices:
          - opamp
          - collector_api
          - inventory
          - kubernetes
      server_url:
        description:
          - OpAMP server URL.
          - Required when type=opamp.
        type: str
      token:
        description:
          - Bearer token for OpAMP authentication.
          - Required when type=opamp.
        type: str
        no_log: true
      hosts:
        description:
          - List of collector host addresses.
          - Required when type=collector_api.
        type: list
        elements: str
      health_port:
        description:
          - Health check port for collector API.
          - Used when type=collector_api.
        type: int
        default: 13133
      group:
        description:
          - Ansible inventory group name.
          - Required when type=inventory.
        type: str
      kubeconfig:
        description:
          - Path to kubeconfig file.
          - Used when type=kubernetes.
          - If not provided, uses in-cluster config.
        type: str
      namespace:
        description:
          - Kubernetes namespace to query.
          - Used when type=kubernetes.
        type: str
        default: default
      label_selector:
        description:
          - Kubernetes label selector for filtering pods.
          - Used when type=kubernetes.
        type: str
  validate_certs:
    description:
      - Whether to validate TLS certificates.
    type: bool
    default: true
  timeout:
    description:
      - Request timeout in seconds.
    type: int
    default: 30
notes:
  - The module deduplicates nodes using host.id as the primary key.
  - When host.id is unavailable, falls back to host.name (generates warning).
  - Discovery source errors are logged as warnings but do not fail the module.
  - For AAP indirect node licensing, use node_summary.total_count.
'''

EXAMPLES = r'''
- name: Count nodes from OpAMP server
  signalfx.splunk_otel_collector.otel_node_count:
    sources:
      - type: opamp
        server_url: https://opamp.example.com:4320
        token: "{{ opamp_token }}"
  register: node_count

- name: Count nodes from multiple sources with deduplication
  signalfx.splunk_otel_collector.otel_node_count:
    sources:
      - type: opamp
        server_url: https://opamp.example.com:4320
        token: "{{ opamp_token }}"
      - type: collector_api
        hosts:
          - web-01.example.com
          - web-02.example.com
          - db-01.example.com
        health_port: 13133
      - type: inventory
        group: otel_collectors
      - type: kubernetes
        namespace: otel-system
        label_selector: app=otel-collector
    timeout: 60
  register: node_count

- name: Display node count summary
  ansible.builtin.debug:
    msg: "Total unique nodes: {{ node_count.node_summary.total_count }}"
'''

RETURN = r'''
nodes:
  description: List of discovered nodes in Azure Resource Manager format
  returned: always
  type: list
  elements: dict
  contains:
    id:
      description: Unique node identifier (host.id)
      type: str
      sample: "machine-id-001"
    name:
      description: Human-readable node name
      type: str
      sample: "web-01"
    location:
      description: Node location or region
      type: str
      sample: "us-east-1"
    tags:
      description: Node tags and labels
      type: dict
      sample: {"env": "production", "region": "us-east-1"}
    properties:
      description: Node properties
      type: dict
      contains:
        host_id:
          description: Machine ID from /etc/machine-id
          type: str
          sample: "machine-id-001"
        host_name:
          description: Hostname
          type: str
          sample: "web-01"
        agent_instance_id:
          description: OpAMP agent instance UUID
          type: str
          sample: "550e8400-e29b-41d4-a716-446655440000"
        collector_version:
          description: OpenTelemetry Collector version
          type: str
          sample: "0.96.0"
        signal_types:
          description: Signal types being collected
          type: list
          elements: str
          sample: ["metrics", "logs", "traces"]
        discovered_via:
          description: List of sources that discovered this node
          type: list
          elements: str
          sample: ["opamp", "collector_api"]
    provisioning_state:
      description: Node provisioning state
      type: str
      sample: "Active"
node_summary:
  description: Aggregated node count summary
  returned: always
  type: dict
  contains:
    total_count:
      description: Total unique nodes (use for AAP licensing)
      type: int
      sample: 47
    unique_host_ids:
      description: Number of unique host.id values
      type: int
      sample: 47
    duplicate_detections:
      description: Number of nodes found in multiple sources
      type: int
      sample: 3
    by_source:
      description: Node count by discovery source
      type: dict
      sample: {"opamp": 45, "collector_api": 47, "inventory": 50}
    by_state:
      description: Node count by provisioning state
      type: dict
      sample: {"active": 45, "unreachable": 2}
warnings:
  description: List of warning messages from discovery
  returned: always
  type: list
  elements: str
  sample:
    - "3 nodes missing host.id, falling back to host.name for dedup"
    - "Collector API discovery failed for web-03: Connection refused"
'''


from ansible.module_utils.basic import AnsibleModule

try:
    from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.node_discovery import NodeDiscovery
except ModuleNotFoundError:
    # Fallback for test environment
    import sys
    import os
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../module_utils'))
    from node_discovery import NodeDiscovery


def run_module():
    """Run the otel_node_count module."""
    # Define module arguments
    source_spec = dict(
        type=dict(type='str', required=True, choices=['opamp', 'collector_api', 'inventory', 'kubernetes']),
        server_url=dict(type='str'),
        token=dict(type='str', no_log=True),
        hosts=dict(type='list', elements='str'),
        health_port=dict(type='int', default=13133),
        group=dict(type='str'),
        kubeconfig=dict(type='str'),
        namespace=dict(type='str', default='default'),
        label_selector=dict(type='str'),
    )

    module_args = dict(
        sources=dict(type='list', required=True, elements='dict', options=source_spec),
        validate_certs=dict(type='bool', default=True),
        timeout=dict(type='int', default=30),
    )

    # Initialize module
    module = AnsibleModule(
        argument_spec=module_args,
        supports_check_mode=True
    )

    # Get parameters
    sources = module.params['sources']
    validate_certs = module.params['validate_certs']
    timeout = module.params['timeout']

    # Create NodeDiscovery instance
    discovery = NodeDiscovery(timeout=timeout, validate_certs=validate_certs)

    # Process each source
    for source in sources:
        source_type = source['type']

        try:
            if source_type == 'opamp':
                # Validate required parameters
                if not source.get('server_url') or not source.get('token'):
                    module.fail_json(
                        msg="opamp source requires 'server_url' and 'token' parameters"
                    )

                discovery.discover_opamp(
                    server_url=source['server_url'],
                    token=source['token']
                )

            elif source_type == 'collector_api':
                # Validate required parameters
                if not source.get('hosts'):
                    module.fail_json(
                        msg="collector_api source requires 'hosts' parameter"
                    )

                discovery.discover_collector_api(
                    hosts=source['hosts'],
                    health_port=source.get('health_port', 13133)
                )

            elif source_type == 'inventory':
                # Get hosts from inventory group
                group = source.get('group')
                if not group:
                    module.fail_json(
                        msg="inventory source requires 'group' parameter"
                    )

                # Get hosts from the specified group
                hosts = []
                if group in module.params.get('groups', {}):
                    hosts = module.params['groups'][group]
                else:
                    # Try to access inventory directly
                    try:
                        hosts = list(module.ansible.inventory.get_hosts(group))
                        hosts = [h.name for h in hosts]
                    except (AttributeError, KeyError):
                        # If we can't access inventory, add a warning
                        discovery._warnings.append(
                            f"Could not access inventory group '{group}'"
                        )
                        hosts = []

                if hosts:
                    discovery.discover_inventory(hosts)

            elif source_type == 'kubernetes':
                discovery.discover_kubernetes(
                    kubeconfig=source.get('kubeconfig'),
                    namespace=source.get('namespace', 'default'),
                    label_selector=source.get('label_selector')
                )

        except Exception as e:
            # Log error as warning but continue
            discovery._warnings.append(
                f"Error processing {source_type} source: {str(e)}"
            )

    # Get results
    results = discovery.get_results()

    # Return results
    module.exit_json(
        changed=False,
        nodes=results['nodes'],
        node_summary=results['node_summary'],
        warnings=results['warnings']
    )


def main():
    """Entry point for module execution."""
    run_module()


if __name__ == '__main__':
    main()
