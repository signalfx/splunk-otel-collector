#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type


DOCUMENTATION = r"""
---
module: opamp_agent_info
short_description: Query OpAMP management server for agent fleet status
version_added: "1.0.0"
description:
  - Queries an OpAMP-compatible management server's administrative API for agent information
  - Returns agent data in Azure resource taxonomy format
  - Supports filtering by agent name and custom tags
author:
  - Cisco Systems (@cisco)
options:
  server_url:
    description:
      - Base URL of the OpAMP management server
      - Should include protocol (http:// or https://)
    type: str
    required: true
  token:
    description:
      - Bearer token for authentication
      - Stored securely and not logged
    type: str
    required: true
    no_log: true
  name:
    description:
      - Filter agents by exact name match
      - Matches against the host.name identifying attribute
    type: str
    required: false
  tags:
    description:
      - Filter agents by custom label tags
      - All specified tags must match (AND logic)
    type: dict
    required: false
  validate_certs:
    description:
      - Whether to validate TLS certificates
    type: bool
    default: true
  timeout:
    description:
      - HTTP request timeout in seconds
    type: int
    default: 30
"""

EXAMPLES = r"""
- name: Get all agents from OpAMP server
  cisco.splunk_otel_collector.opamp_agent_info:
    server_url: "https://opamp.example.com:4320"
    token: "{{ opamp_token }}"
  register: all_agents

- name: Get agents in production rollout group
  cisco.splunk_otel_collector.opamp_agent_info:
    server_url: "https://opamp.example.com:4320"
    token: "{{ opamp_token }}"
    tags:
      rollout_group: "production"
  register: prod_agents

- name: Get specific agent by name
  cisco.splunk_otel_collector.opamp_agent_info:
    server_url: "https://opamp.example.com:4320"
    token: "{{ opamp_token }}"
    name: "web-01"
  register: web_agent

- name: Get agents in us-east-1 region with staging rollout
  cisco.splunk_otel_collector.opamp_agent_info:
    server_url: "https://opamp.example.com:4320"
    token: "{{ opamp_token }}"
    tags:
      region: "us-east-1"
      rollout_group: "staging"
  register: filtered_agents
"""

RETURN = r"""
agents:
  description: List of agents in Azure resource taxonomy format
  returned: always
  type: list
  elements: dict
  sample:
    - id: "opamp://opamp.example.com:4320/550e8400-e29b-41d4-a716-446655440000"
      name: "web-01"
      location: "web-01"
      tags:
        rollout_group: "production"
        region: "us-east-1"
      properties:
        agent_version: "0.96.0"
        capabilities: ["AcceptsRemoteConfig", "ReportsHealth"]
        effective_config_hash: "abc123def456"
        last_heartbeat: ""
        health_status: "healthy"
        service_name: "otel-collector"
      provisioning_state: "Connected"
"""

from typing import TYPE_CHECKING

from ansible.module_utils.basic import AnsibleModule

if TYPE_CHECKING:
    from typing import Any, Dict, List, Optional

try:
    from ansible_collections.cisco.splunk_otel_collector.plugins.module_utils.opamp_client import OpAMPClient
    HAS_OPAMP_CLIENT = True
except ImportError:
    try:
        from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.opamp_client import OpAMPClient
        HAS_OPAMP_CLIENT = True
    except ImportError:
        HAS_OPAMP_CLIENT = False


def run_module() -> None:
    """Main module execution."""
    module_args = dict(
        server_url=dict(type='str', required=True),
        token=dict(type='str', required=True, no_log=True),
        name=dict(type='str', required=False),
        tags=dict(type='dict', required=False),
        validate_certs=dict(type='bool', default=True),
        timeout=dict(type='int', default=30),
    )

    result: Dict[str, Any] = dict(
        changed=False,
        agents=[],
    )

    module = AnsibleModule(
        argument_spec=module_args,
        supports_check_mode=True
    )

    if not HAS_OPAMP_CLIENT:
        module.fail_json(
            msg='opamp_client module_utils not found',
            **result
        )

    # Extract parameters
    server_url = module.params['server_url']
    token = module.params['token']
    name = module.params.get('name')
    tags = module.params.get('tags')
    validate_certs = module.params['validate_certs']
    timeout = module.params['timeout']

    try:
        # Initialize client and query agents
        client = OpAMPClient(
            server_url=server_url,
            token=token,
            validate_certs=validate_certs,
            timeout=timeout
        )

        agents = client.get_agents()

        # Apply filters
        if name is not None or tags is not None:
            agents = client.filter_agents(agents, name=name, tags=tags)

        result['agents'] = agents

    except Exception as e:
        module.fail_json(
            msg=f'Failed to query OpAMP server: {str(e)}',
            **result
        )

    module.exit_json(**result)


def main() -> None:
    """Entry point for module execution."""
    run_module()


if __name__ == '__main__':
    main()
