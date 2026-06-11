#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

DOCUMENTATION = r"""
---
name: opamp
plugin_type: inventory
short_description: Dynamic inventory from OpAMP management server
version_added: "1.0.0"
description:
  - Generates Ansible inventory from an OpAMP management server
  - Queries the server's administrative API for agent fleet status
  - Creates host entries for each agent with metadata as host variables
  - Supports grouping by rollout group, region, agent version, and health status
author:
  - Cisco Systems (@cisco)
options:
  plugin:
    description:
      - Token identifying this as the opamp inventory plugin
      - Must be 'cisco.splunk_otel_collector.opamp' or 'signalfx.splunk_otel_collector.opamp'
    required: true
    choices: ['cisco.splunk_otel_collector.opamp', 'signalfx.splunk_otel_collector.opamp']
  server_url:
    description:
      - Base URL of the OpAMP management server
      - Should include protocol (http:// or https://)
    type: str
    required: true
  token:
    description:
      - Bearer token for authentication
      - Can also be set via OPAMP_TOKEN environment variable
    type: str
    required: true
    env:
      - name: OPAMP_TOKEN
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
  strict:
    description:
      - If C(yes), make invalid entries a fatal error
      - Otherwise, skip and continue
    type: bool
    default: false
  compose:
    description:
      - Create host variables from Jinja2 expressions
    type: dict
    default: {}
  groups:
    description:
      - Add hosts to groups based on Jinja2 conditionals
    type: dict
    default: {}
  keyed_groups:
    description:
      - Add hosts to groups based on variable values
      - Supports grouping by any agent property or tag
    type: list
    elements: dict
    default: []
"""

EXAMPLES = r"""
# inventory.opamp.yml
plugin: cisco.splunk_otel_collector.opamp
server_url: https://opamp.example.com:4320
token: "{{ lookup('env', 'OPAMP_TOKEN') }}"
validate_certs: true

# Group by rollout_group tag
keyed_groups:
  - key: rollout_group
    prefix: rollout
    separator: "_"

# Group by region tag
  - key: region
    prefix: region
    separator: "_"

# Group by agent version
  - key: agent_version
    prefix: version
    separator: "_"

# Group by health status
  - key: health_status
    prefix: health
    separator: "_"

# Create custom variables
compose:
  ansible_host: name
  collector_version: agent_version

# Create custom groups
groups:
  production: rollout_group == "production"
  staging: rollout_group == "staging"
  unhealthy: health_status == "unhealthy"
"""

import os
from typing import TYPE_CHECKING, Any, Dict, List

from ansible.errors import AnsibleError
from ansible.plugins.inventory import BaseInventoryPlugin, Constructable, Cacheable

if TYPE_CHECKING:
    from ansible.inventory.data import InventoryData

try:
    from ansible_collections.cisco.splunk_otel_collector.plugins.module_utils.opamp_client import OpAMPClient
    HAS_OPAMP_CLIENT = True
except ImportError:
    try:
        from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.opamp_client import OpAMPClient
        HAS_OPAMP_CLIENT = True
    except ImportError:
        HAS_OPAMP_CLIENT = False


class InventoryModule(BaseInventoryPlugin, Constructable, Cacheable):
    """OpAMP dynamic inventory plugin."""

    NAME = 'signalfx.splunk_otel_collector.opamp'

    def verify_file(self, path: str) -> bool:
        """Verify this is a valid inventory file.

        Args:
            path: Path to inventory file

        Returns:
            True if file is valid for this plugin
        """
        valid = False
        if super().verify_file(path):
            if path.endswith(('opamp.yml', 'opamp.yaml')):
                valid = True
        return valid

    def parse(
        self,
        inventory: 'InventoryData',
        loader: Any,
        path: str,
        cache: bool = True
    ) -> None:
        """Parse inventory and populate inventory object.

        Args:
            inventory: Inventory object to populate
            loader: Data loader
            path: Path to inventory file
            cache: Whether to use cache
        """
        super().parse(inventory, loader, path, cache)

        if not HAS_OPAMP_CLIENT:
            raise AnsibleError('opamp_client module_utils not found')

        # Read configuration
        self._read_config_data(path)

        # Get config options
        server_url = self.get_option('server_url')
        token = self.get_option('token')
        validate_certs = self.get_option('validate_certs')
        timeout = self.get_option('timeout')
        strict = self.get_option('strict')

        # Query OpAMP server
        try:
            client = OpAMPClient(
                server_url=server_url,
                token=token,
                validate_certs=validate_certs,
                timeout=timeout
            )
            agents = client.get_agents()
        except Exception as e:
            raise AnsibleError(f'Failed to query OpAMP server: {str(e)}')

        # Populate inventory
        for agent in agents:
            self._populate_host(inventory, agent, strict)

    def _populate_host(
        self,
        inventory: 'InventoryData',
        agent: Dict[str, Any],
        strict: bool
    ) -> None:
        """Add agent as host to inventory.

        Args:
            inventory: Inventory object
            agent: Agent data dict
            strict: Whether to raise errors or skip
        """
        # Use agent name as hostname
        hostname = agent['name']

        # Add host
        self.inventory.add_host(hostname)

        # Set basic host variables
        self.inventory.set_variable(hostname, 'ansible_host', agent['name'])
        self.inventory.set_variable(hostname, 'agent_id', agent['id'])
        self.inventory.set_variable(hostname, 'location', agent['location'])
        self.inventory.set_variable(hostname, 'provisioning_state', agent['provisioning_state'])

        # Set properties as host variables
        for key, value in agent['properties'].items():
            self.inventory.set_variable(hostname, key, value)

        # Set tags as host variables
        for key, value in agent['tags'].items():
            self.inventory.set_variable(hostname, key, value)

        # Build variables dict for compose/groups/keyed_groups
        variables = {
            'ansible_host': agent['name'],
            'agent_id': agent['id'],
            'location': agent['location'],
            'provisioning_state': agent['provisioning_state'],
        }
        variables.update(agent['properties'])
        variables.update(agent['tags'])

        # Apply compose, groups, and keyed_groups
        self._set_composite_vars(
            self.get_option('compose'),
            variables,
            hostname,
            strict=strict
        )

        self._add_host_to_composed_groups(
            self.get_option('groups'),
            variables,
            hostname,
            strict=strict
        )

        self._add_host_to_keyed_groups(
            self.get_option('keyed_groups'),
            variables,
            hostname,
            strict=strict
        )
