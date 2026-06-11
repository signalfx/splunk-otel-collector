#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import json
import os
import sys
from unittest.mock import MagicMock, patch, mock_open
from urllib.error import HTTPError, URLError

import pytest

# Add module paths
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))

from node_discovery import NodeDiscovery


@pytest.fixture
def discovery():
    """Fixture providing a NodeDiscovery instance."""
    return NodeDiscovery(timeout=30, validate_certs=True)


def test_initialization():
    """Test NodeDiscovery initialization."""
    disc = NodeDiscovery(timeout=60, validate_certs=False)
    assert disc._timeout == 60
    assert disc._validate_certs is False
    assert disc._nodes == {}
    assert disc._source_counts == {}
    assert disc._warnings == []
    assert disc._fallback_count == 0


def test_merge_single_node(discovery):
    """Test merging a single node."""
    discovery._merge_node(
        host_id='machine-id-001',
        host_name='web-01',
        source='opamp',
        location='us-east-1',
        tags={'env': 'prod'},
        collector_version='0.96.0'
    )

    assert len(discovery._nodes) == 1
    node = discovery._nodes['machine-id-001']
    assert node['id'] == 'machine-id-001'
    assert node['name'] == 'web-01'
    assert node['location'] == 'us-east-1'
    assert node['tags'] == {'env': 'prod'}
    assert node['properties']['host_id'] == 'machine-id-001'
    assert node['properties']['host_name'] == 'web-01'
    assert node['properties']['discovered_via'] == ['opamp']
    assert node['properties']['collector_version'] == '0.96.0'


def test_merge_duplicate_host_id(discovery):
    """Test merging nodes with same host_id from different sources."""
    # First discovery from opamp
    discovery._merge_node(
        host_id='machine-id-001',
        host_name='web-01',
        source='opamp',
        location='us-east-1',
        tags={'env': 'prod'},
        collector_version='0.96.0'
    )

    # Second discovery from collector_api
    discovery._merge_node(
        host_id='machine-id-001',
        host_name='web-01',
        source='collector_api',
        location='us-east-1',
        tags={'region': 'east'},
        provisioning_state='Active'
    )

    # Should only have one node
    assert len(discovery._nodes) == 1
    node = discovery._nodes['machine-id-001']

    # discovered_via should be a list with both sources
    assert set(node['properties']['discovered_via']) == {'opamp', 'collector_api'}

    # Tags should be merged
    assert node['tags']['env'] == 'prod'
    assert node['tags']['region'] == 'east'

    # provisioning_state should be updated
    assert node['provisioning_state'] == 'Active'


def test_fallback_to_hostname(discovery):
    """Test fallback to hostname when no host.id is available."""
    discovery._merge_node(
        host_id='web-01',  # Same as host_name
        host_name='web-01',
        source='collector_api',
        location='web-01',
        tags={},
        provisioning_state='Active'
    )

    # Should have incremented fallback count
    assert discovery._fallback_count == 1

    results = discovery.get_results()
    # Should have a warning about missing host.id
    assert any('missing host.id' in w for w in results['warnings'])


def test_discover_inventory(discovery):
    """Test discovering nodes from inventory."""
    hosts = ['web-01', 'web-02', 'db-01']
    discovery.discover_inventory(hosts)

    assert discovery._source_counts['inventory'] == 3
    assert len(discovery._nodes) == 3

    # Check one of the nodes
    node = discovery._nodes['web-01']
    assert node['name'] == 'web-01'
    assert node['properties']['discovered_via'] == ['inventory']
    assert node['provisioning_state'] == 'Active'


@patch('node_discovery.urlopen')
def test_discover_collector_api_success(mock_urlopen, discovery):
    """Test successful collector API discovery."""
    # Mock successful health check responses
    mock_response = MagicMock()
    mock_response.read.return_value = b'OK'
    mock_urlopen.return_value = mock_response

    hosts = ['web-01', 'web-02']
    discovery.discover_collector_api(hosts, health_port=13133)

    # Should have called urlopen for each host
    assert mock_urlopen.call_count == 2

    # Should have discovered 2 nodes
    assert discovery._source_counts['collector_api'] == 2
    assert len(discovery._nodes) == 2

    # Check node properties
    node = discovery._nodes['web-01']
    assert node['name'] == 'web-01'
    assert node['properties']['discovered_via'] == ['collector_api']
    assert node['provisioning_state'] == 'Active'


@patch('node_discovery.urlopen')
def test_discover_collector_api_unreachable(mock_urlopen, discovery):
    """Test collector API discovery with unreachable host."""
    # Mock connection error
    mock_urlopen.side_effect = URLError('Connection refused')

    hosts = ['web-01']
    discovery.discover_collector_api(hosts, health_port=13133)

    # Should still have discovered the node but marked as unreachable
    assert discovery._source_counts['collector_api'] == 1
    assert len(discovery._nodes) == 1

    node = discovery._nodes['web-01']
    assert node['provisioning_state'] == 'Unreachable'

    # Should have a warning
    assert len(discovery._warnings) > 0
    assert 'web-01' in discovery._warnings[0]


def test_discover_opamp(discovery):
    """Test OpAMP discovery."""
    # Import and patch OpAMPClient
    with patch('opamp_client.OpAMPClient') as mock_opamp_client_class:
        # Mock OpAMPClient
        mock_client = MagicMock()
        mock_opamp_client_class.return_value = mock_client

        # Mock agent data
        mock_client.get_agents.return_value = [
            {
                'id': 'opamp://server/instance-001',
                'name': 'web-01',
                'location': 'us-east-1',
                'tags': {'env': 'prod'},
                'properties': {
                    'host_id': 'machine-id-001',
                    'agent_version': '0.96.0',
                    'health_status': 'healthy'
                },
                'provisioning_state': 'Connected'
            },
            {
                'id': 'opamp://server/instance-002',
                'name': 'web-02',
                'location': 'us-west-1',
                'tags': {'env': 'staging'},
                'properties': {
                    'host_id': 'machine-id-002',
                    'agent_version': '0.95.0',
                    'health_status': 'unhealthy'
                },
                'provisioning_state': 'Disconnected'
            }
        ]

        discovery.discover_opamp('https://opamp.example.com', 'test-token')

        # Should have created OpAMPClient
        mock_opamp_client_class.assert_called_once_with(
            server_url='https://opamp.example.com',
            token='test-token',
            validate_certs=True,
            timeout=30
        )

        # Should have discovered 2 nodes
        assert discovery._source_counts['opamp'] == 2
        assert len(discovery._nodes) == 2

        # Check first node
        node = discovery._nodes['machine-id-001']
        assert node['name'] == 'web-01'
        assert node['properties']['host_id'] == 'machine-id-001'
        assert node['properties']['collector_version'] == '0.96.0'
        assert node['properties']['discovered_via'] == ['opamp']


def test_discover_opamp_error(discovery):
    """Test OpAMP discovery with error."""
    with patch('opamp_client.OpAMPClient') as mock_opamp_client_class:
        # Mock OpAMPClient that raises error
        mock_client = MagicMock()
        mock_opamp_client_class.return_value = mock_client
        mock_client.get_agents.side_effect = HTTPError(
            'https://opamp.example.com/v1/agents',
            401,
            'Unauthorized',
            {},
            None
        )

        discovery.discover_opamp('https://opamp.example.com', 'bad-token')

        # Should have handled error gracefully
        assert discovery._source_counts['opamp'] == 0
        assert len(discovery._warnings) > 0
        assert 'OpAMP discovery failed' in discovery._warnings[0]


@patch('node_discovery.urlopen')
def test_discover_kubernetes(mock_urlopen, discovery):
    """Test Kubernetes discovery."""
    # Mock K8s API response
    k8s_response = {
        'items': [
            {
                'metadata': {'name': 'collector-pod-1'},
                'spec': {'nodeName': 'k8s-node-01'}
            },
            {
                'metadata': {'name': 'collector-pod-2'},
                'spec': {'nodeName': 'k8s-node-02'}
            },
            {
                'metadata': {'name': 'collector-pod-3'},
                'spec': {'nodeName': 'k8s-node-01'}  # Duplicate node
            }
        ]
    }

    mock_response = MagicMock()
    mock_response.read.return_value = json.dumps(k8s_response).encode('utf-8')
    mock_urlopen.return_value = mock_response

    # Mock in-cluster token file
    with patch('builtins.open', mock_open(read_data='test-token')):
        discovery.discover_kubernetes(
            kubeconfig=None,
            namespace='otel',
            label_selector='app=otel-collector'
        )

    # Should have discovered 2 unique nodes (k8s-node-01, k8s-node-02)
    assert discovery._source_counts['kubernetes'] == 2
    assert len(discovery._nodes) == 2

    # Check nodes
    assert 'k8s-node-01' in discovery._nodes
    assert 'k8s-node-02' in discovery._nodes

    node = discovery._nodes['k8s-node-01']
    assert node['tags']['namespace'] == 'otel'


@patch('node_discovery.urlopen')
def test_discover_kubernetes_error(mock_urlopen, discovery):
    """Test Kubernetes discovery with API error."""
    mock_urlopen.side_effect = URLError('Connection refused')

    with patch('builtins.open', mock_open(read_data='test-token')):
        discovery.discover_kubernetes(namespace='default')

    # Should have handled error gracefully
    assert discovery._source_counts['kubernetes'] == 0
    assert len(discovery._warnings) > 0
    assert 'Kubernetes discovery failed' in discovery._warnings[0]


def test_get_results_summary(discovery):
    """Test get_results summary generation."""
    # Add some nodes from different sources
    discovery._merge_node(
        host_id='machine-id-001',
        host_name='web-01',
        source='opamp',
        provisioning_state='Active'
    )
    discovery._merge_node(
        host_id='machine-id-001',
        host_name='web-01',
        source='collector_api',
        provisioning_state='Active'
    )
    discovery._merge_node(
        host_id='machine-id-002',
        host_name='web-02',
        source='opamp',
        provisioning_state='Unreachable'
    )

    discovery._source_counts = {'opamp': 2, 'collector_api': 1}

    results = discovery.get_results()

    # Check structure
    assert 'nodes' in results
    assert 'node_summary' in results
    assert 'warnings' in results

    # Check node_summary
    summary = results['node_summary']
    assert summary['total_count'] == 2
    assert summary['unique_host_ids'] == 2
    assert summary['duplicate_detections'] == 1  # machine-id-001 from 2 sources
    assert summary['by_source'] == {'opamp': 2, 'collector_api': 1}
    assert summary['by_state'] == {'active': 1, 'unreachable': 1}


def test_empty_sources(discovery):
    """Test get_results with no sources."""
    results = discovery.get_results()

    assert results['nodes'] == []
    assert results['node_summary']['total_count'] == 0
    assert results['node_summary']['unique_host_ids'] == 0
    assert results['node_summary']['duplicate_detections'] == 0
    assert results['node_summary']['by_source'] == {}
    assert results['node_summary']['by_state'] == {}


def test_extract_instance_id(discovery):
    """Test extracting instance ID from agent."""
    agent = {'id': 'opamp://server.example.com/550e8400-e29b-41d4-a716-446655440000'}
    instance_id = discovery._extract_instance_id(agent)
    assert instance_id == '550e8400-e29b-41d4-a716-446655440000'

    # Test with simple ID
    agent = {'id': 'simple-id'}
    instance_id = discovery._extract_instance_id(agent)
    assert instance_id == 'simple-id'
