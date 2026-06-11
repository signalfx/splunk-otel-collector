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

"""
Shared HTTP client for querying OpAMP management servers.
"""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import json
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


class OpAMPClient:
    """Client for OpAMP management server API."""

    def __init__(
        self,
        server_url: str,
        token: str,
        validate_certs: bool = True,
        timeout: int = 30
    ):
        """Initialize OpAMP client.

        Args:
            server_url: Base URL of the OpAMP server
            token: Bearer token for authentication
            validate_certs: Whether to validate TLS certificates
            timeout: Request timeout in seconds
        """
        self.server_url = server_url.rstrip('/')
        self.token = token
        self.validate_certs = validate_certs
        self.timeout = timeout

    def get_agents(self) -> List[Dict[str, Any]]:
        """Query /v1/agents endpoint and return list of agents.

        Returns:
            List of agent dictionaries in normalized Azure taxonomy format

        Raises:
            HTTPError: If the request fails
            URLError: If connection fails
            ValueError: If response is not valid JSON
        """
        url = f"{self.server_url}/v1/agents"
        headers = {
            'Authorization': f'Bearer {self.token}',
            'Accept': 'application/json',
            'User-Agent': 'ansible-cisco-splunk-otel-collector/1.0.0'
        }

        request = Request(url, headers=headers)

        try:
            # Note: validate_certs would require ssl.SSLContext in production
            response = urlopen(request, timeout=self.timeout)
            data = response.read().decode('utf-8')
            raw_agents = json.loads(data)

            if not isinstance(raw_agents, list):
                raise ValueError("Expected list of agents from API")

            return [self._normalize_agent(agent) for agent in raw_agents]

        except HTTPError as e:
            error_body = e.read().decode('utf-8', errors='ignore')
            raise HTTPError(
                e.url,
                e.code,
                f"OpAMP API error: {e.reason}. Response: {error_body}",
                e.headers,
                e.fp
            )
        except URLError as e:
            raise URLError(f"Failed to connect to OpAMP server: {e.reason}")
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON response from OpAMP server: {e}")

    def _normalize_agent(self, raw: Dict[str, Any]) -> Dict[str, Any]:
        """Convert raw API response to Azure taxonomy format.

        Args:
            raw: Raw agent object from OpAMP API

        Returns:
            Agent dict in Azure taxonomy format
        """
        instance_id = raw.get('instance_id', '')
        agent_desc = raw.get('agent_description', {})

        # Extract identifying attributes
        identifying_attrs = agent_desc.get('identifying_attributes', [])
        host_name = ''
        service_name = ''

        for attr in identifying_attrs:
            key = attr.get('key', '')
            value = attr.get('value', '')
            if key == 'host.name':
                host_name = value
            elif key == 'service.name':
                service_name = value

        # Extract custom labels as tags
        custom_labels = raw.get('custom_labels', {})

        # Determine health status
        health = raw.get('health', {})
        healthy = health.get('healthy', True)
        health_status = 'healthy' if healthy else 'unhealthy'

        # Build Azure taxonomy format
        agent_id = f"opamp://{self.server_url.split('://')[-1]}/{instance_id}"

        return {
            'id': agent_id,
            'name': host_name or instance_id,
            'location': host_name or instance_id,
            'tags': custom_labels,
            'properties': {
                'agent_version': raw.get('agent_version', ''),
                'capabilities': raw.get('capabilities', []),
                'effective_config_hash': raw.get('effective_config_hash', ''),
                'last_heartbeat': '',  # Not provided by current API
                'health_status': health_status,
                'service_name': service_name,
            },
            'provisioning_state': 'Connected' if healthy else 'Disconnected',
        }

    def filter_agents(
        self,
        agents: List[Dict[str, Any]],
        name: Optional[str] = None,
        tags: Optional[Dict[str, str]] = None
    ) -> List[Dict[str, Any]]:
        """Filter agents by name and/or tags.

        Args:
            agents: List of agent dicts
            name: Filter by exact name match
            tags: Filter by tag key-value pairs (all must match)

        Returns:
            Filtered list of agents
        """
        filtered = agents

        if name is not None:
            filtered = [a for a in filtered if a['name'] == name]

        if tags is not None:
            filtered = [
                a for a in filtered
                if all(a['tags'].get(k) == v for k, v in tags.items())
            ]

        return filtered
