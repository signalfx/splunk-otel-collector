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
Node discovery utility for aggregating and deduplicating nodes from multiple sources.
"""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import json
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


class NodeDiscovery:
    """Discover and aggregate nodes from multiple sources with deduplication."""

    def __init__(self, timeout: int = 30, validate_certs: bool = True):
        """Initialize NodeDiscovery.

        Args:
            timeout: Request timeout in seconds
            validate_certs: Whether to validate TLS certificates
        """
        self._nodes: Dict[str, Dict[str, Any]] = {}  # keyed by host_id
        self._source_counts: Dict[str, int] = {}
        self._warnings: List[str] = []
        self._fallback_count = 0
        self._timeout = timeout
        self._validate_certs = validate_certs

    def discover_opamp(self, server_url: str, token: str) -> None:
        """Query OpAMP server and extract nodes.

        Args:
            server_url: Base URL of the OpAMP server
            token: Bearer token for authentication
        """
        try:
            from ansible_collections.signalfx.splunk_otel_collector.plugins.module_utils.opamp_client import OpAMPClient
        except ModuleNotFoundError:
            # Fallback for test environment
            from opamp_client import OpAMPClient

        try:
            client = OpAMPClient(
                server_url=server_url,
                token=token,
                validate_certs=self._validate_certs,
                timeout=self._timeout
            )
            agents = client.get_agents()

            for agent in agents:
                # Extract host.id from identifying attributes
                # OpAMP stores this in properties or we derive from agent data
                host_id = self._extract_host_id_from_agent(agent)
                host_name = agent.get('name', '')

                # Extract additional properties
                props = agent.get('properties', {})
                signal_types = self._infer_signal_types(agent)

                self._merge_node(
                    host_id=host_id,
                    host_name=host_name,
                    source='opamp',
                    location=agent.get('location', ''),
                    tags=agent.get('tags', {}),
                    agent_instance_id=self._extract_instance_id(agent),
                    collector_version=props.get('agent_version', ''),
                    signal_types=signal_types,
                    provisioning_state=agent.get('provisioning_state', 'Unknown')
                )

            self._source_counts['opamp'] = len(agents)

        except (HTTPError, URLError, ValueError) as e:
            self._warnings.append(f"OpAMP discovery failed: {str(e)}")
            self._source_counts['opamp'] = 0

    def discover_collector_api(self, hosts: List[str], health_port: int = 13133) -> None:
        """Query each collector health endpoint.

        Args:
            hosts: List of host addresses
            health_port: Health check port (default 13133)
        """
        discovered = 0

        for host in hosts:
            try:
                url = f"http://{host}:{health_port}/"
                request = Request(url)
                response = urlopen(request, timeout=self._timeout)

                # Collector is reachable
                # Try to extract host.id if available (some collectors expose metadata)
                # For now, use hostname as identifier
                host_id = host
                host_name = host

                self._merge_node(
                    host_id=host_id,
                    host_name=host_name,
                    source='collector_api',
                    location=host,
                    tags={},
                    provisioning_state='Active'
                )
                discovered += 1

            except (HTTPError, URLError, OSError) as e:
                # Host unreachable or unhealthy
                self._merge_node(
                    host_id=host,
                    host_name=host,
                    source='collector_api',
                    location=host,
                    tags={},
                    provisioning_state='Unreachable'
                )
                self._warnings.append(f"Collector API discovery failed for {host}: {str(e)}")
                discovered += 1

        self._source_counts['collector_api'] = discovered

    def discover_inventory(self, group_hosts: List[str]) -> None:
        """Count hosts from inventory group.

        Args:
            group_hosts: List of hostnames from inventory
        """
        for host_name in group_hosts:
            # Use hostname as identifier (no host.id available from inventory)
            self._merge_node(
                host_id=host_name,
                host_name=host_name,
                source='inventory',
                location=host_name,
                tags={},
                provisioning_state='Active'
            )

        self._source_counts['inventory'] = len(group_hosts)

    def discover_kubernetes(
        self,
        kubeconfig: Optional[str] = None,
        namespace: str = 'default',
        label_selector: Optional[str] = None
    ) -> None:
        """Query K8s API for collector pods.

        Args:
            kubeconfig: Path to kubeconfig (optional, uses in-cluster if not provided)
            namespace: Kubernetes namespace
            label_selector: Label selector for filtering pods
        """
        try:
            # Attempt to get K8s API token
            token = self._get_k8s_token(kubeconfig)
            api_server = self._get_k8s_api_server(kubeconfig)

            # Build API URL
            url = f"{api_server}/api/v1/namespaces/{namespace}/pods"
            if label_selector:
                url += f"?labelSelector={label_selector}"

            headers = {
                'Authorization': f'Bearer {token}',
                'Accept': 'application/json'
            }

            request = Request(url, headers=headers)
            response = urlopen(request, timeout=self._timeout)
            data = json.loads(response.read().decode('utf-8'))

            # Extract unique node names from pods
            nodes_seen = set()
            for pod in data.get('items', []):
                node_name = pod.get('spec', {}).get('nodeName')
                if node_name and node_name not in nodes_seen:
                    nodes_seen.add(node_name)

                    self._merge_node(
                        host_id=node_name,
                        host_name=node_name,
                        source='kubernetes',
                        location=node_name,
                        tags={'namespace': namespace},
                        provisioning_state='Active'
                    )

            self._source_counts['kubernetes'] = len(nodes_seen)

        except Exception as e:
            self._warnings.append(f"Kubernetes discovery failed: {str(e)}")
            self._source_counts['kubernetes'] = 0

    def _merge_node(
        self,
        host_id: str,
        host_name: str,
        source: str,
        **properties
    ) -> None:
        """Merge a discovered node, deduplicating by host_id.

        Args:
            host_id: Unique host identifier
            host_name: Human-readable hostname
            source: Discovery source name
            **properties: Additional node properties
        """
        # Track if this is a fallback (no real host.id)
        if host_id == host_name and source != 'inventory':
            # This is a fallback case where we don't have a true host.id
            self._fallback_count += 1

        if host_id in self._nodes:
            # Node already exists, merge sources
            existing = self._nodes[host_id]
            discovered_via = existing['properties'].get('discovered_via', [])
            if isinstance(discovered_via, str):
                discovered_via = [discovered_via]
            if source not in discovered_via:
                discovered_via.append(source)
            existing['properties']['discovered_via'] = discovered_via

            # Update provisioning_state if new state is more specific
            if properties.get('provisioning_state'):
                existing['provisioning_state'] = properties['provisioning_state']

            # Merge tags
            existing_tags = existing.get('tags', {})
            new_tags = properties.get('tags', {})
            existing_tags.update(new_tags)
            existing['tags'] = existing_tags

            # Update other properties
            for key, value in properties.items():
                if key not in ('discovered_via', 'tags', 'provisioning_state') and value:
                    existing['properties'][key] = value

        else:
            # New node
            self._nodes[host_id] = {
                'id': host_id,
                'name': host_name,
                'location': properties.get('location', host_name),
                'tags': properties.get('tags', {}),
                'properties': {
                    'host_id': host_id,
                    'host_name': host_name,
                    'discovered_via': [source],
                },
                'provisioning_state': properties.get('provisioning_state', 'Unknown')
            }

            # Add additional properties
            for key, value in properties.items():
                if key not in ('location', 'tags', 'provisioning_state') and value:
                    self._nodes[host_id]['properties'][key] = value

    def get_results(self) -> Dict[str, Any]:
        """Return nodes list, node_summary, and warnings.

        Returns:
            Dictionary with nodes, node_summary, and warnings
        """
        nodes_list = list(self._nodes.values())

        # Count by state
        by_state = {}
        for node in nodes_list:
            state = node.get('provisioning_state', 'Unknown')
            by_state[state.lower()] = by_state.get(state.lower(), 0) + 1

        # Track duplicate detections (nodes discovered from multiple sources)
        duplicate_detections = sum(
            1 for node in nodes_list
            if len(node['properties'].get('discovered_via', [])) > 1
        )

        node_summary = {
            'total_count': len(nodes_list),
            'unique_host_ids': len(nodes_list),
            'duplicate_detections': duplicate_detections,
            'by_source': self._source_counts.copy(),
            'by_state': by_state
        }

        # Add warning if we had to fall back to hostnames
        if self._fallback_count > 0:
            self._warnings.append(
                f"{self._fallback_count} nodes missing host.id, falling back to host.name for dedup"
            )

        return {
            'nodes': nodes_list,
            'node_summary': node_summary,
            'warnings': self._warnings
        }

    def _extract_host_id_from_agent(self, agent: Dict[str, Any]) -> str:
        """Extract host.id from agent data.

        Args:
            agent: Agent dictionary

        Returns:
            Host ID or fallback to host name
        """
        # Try to get host.id from properties
        props = agent.get('properties', {})
        if 'host_id' in props:
            return props['host_id']

        # Fallback to using name
        return agent.get('name', agent.get('id', ''))

    def _extract_instance_id(self, agent: Dict[str, Any]) -> str:
        """Extract agent instance ID.

        Args:
            agent: Agent dictionary

        Returns:
            Instance ID
        """
        # Extract from agent ID (format: opamp://server/instance-id)
        agent_id = agent.get('id', '')
        if '://' in agent_id and '/' in agent_id:
            return agent_id.split('/')[-1]
        return agent_id

    def _infer_signal_types(self, agent: Dict[str, Any]) -> List[str]:
        """Infer signal types from agent capabilities.

        Args:
            agent: Agent dictionary

        Returns:
            List of signal types
        """
        # This is a placeholder - actual logic would inspect agent config
        # For now, assume metrics and logs are common
        return ['metrics', 'logs']

    def _get_k8s_token(self, kubeconfig: Optional[str]) -> str:
        """Get Kubernetes API token.

        Args:
            kubeconfig: Path to kubeconfig file

        Returns:
            Bearer token

        Raises:
            ValueError: If token cannot be found
        """
        if kubeconfig:
            # Parse kubeconfig file
            with open(kubeconfig, 'r') as f:
                config = json.load(f)
            # Simplified - in reality, kubeconfig parsing is more complex
            # This is a placeholder
            raise ValueError("Kubeconfig parsing not yet implemented")
        else:
            # Try in-cluster config
            try:
                with open('/var/run/secrets/kubernetes.io/serviceaccount/token', 'r') as f:
                    return f.read().strip()
            except (OSError, IOError) as e:
                raise ValueError(f"Cannot read in-cluster token: {e}")

    def _get_k8s_api_server(self, kubeconfig: Optional[str]) -> str:
        """Get Kubernetes API server URL.

        Args:
            kubeconfig: Path to kubeconfig file

        Returns:
            API server URL

        Raises:
            ValueError: If API server cannot be determined
        """
        if kubeconfig:
            # Parse kubeconfig file
            raise ValueError("Kubeconfig parsing not yet implemented")
        else:
            # In-cluster default
            return "https://kubernetes.default.svc"
