#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import json
import os
import sys
from unittest.mock import MagicMock, patch
from urllib.error import HTTPError, URLError

import pytest

# Add module paths
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))

from opamp_client import OpAMPClient


# Sample raw API response
SAMPLE_RAW_RESPONSE = [
    {
        'instance_id': '550e8400-e29b-41d4-a716-446655440000',
        'agent_description': {
            'identifying_attributes': [
                {'key': 'host.name', 'value': 'web-01'},
                {'key': 'service.name', 'value': 'otel-collector'}
            ],
            'non_identifying_attributes': [
                {'key': 'os.type', 'value': 'linux'}
            ]
        },
        'agent_version': '0.96.0',
        'capabilities': ['AcceptsRemoteConfig', 'ReportsHealth'],
        'effective_config_hash': 'abc123def456',
        'health': {'healthy': True, 'last_error': ''},
        'custom_labels': {'rollout_group': 'production', 'region': 'us-east-1'}
    },
    {
        'instance_id': '660e8400-e29b-41d4-a716-446655440001',
        'agent_description': {
            'identifying_attributes': [
                {'key': 'host.name', 'value': 'web-02'},
                {'key': 'service.name', 'value': 'otel-collector'}
            ],
            'non_identifying_attributes': []
        },
        'agent_version': '0.95.0',
        'capabilities': ['AcceptsRemoteConfig'],
        'effective_config_hash': 'xyz789abc123',
        'health': {'healthy': False, 'last_error': 'connection timeout'},
        'custom_labels': {'rollout_group': 'staging', 'region': 'us-west-2'}
    }
]


@pytest.fixture
def opamp_client():
    """Fixture providing an OpAMPClient instance."""
    return OpAMPClient(
        server_url='https://opamp.example.com:4320',
        token='test-token',
        validate_certs=True,
        timeout=30
    )


def test_client_initialization():
    """Test OpAMPClient initialization."""
    client = OpAMPClient(
        server_url='https://opamp.example.com:4320/',
        token='my-token',
        validate_certs=False,
        timeout=60
    )

    assert client.server_url == 'https://opamp.example.com:4320'
    assert client.token == 'my-token'
    assert client.validate_certs is False
    assert client.timeout == 60


@patch('opamp_client.urlopen')
def test_get_agents_success(mock_urlopen, opamp_client):
    """Test successful agent query."""
    # Mock HTTP response
    mock_response = MagicMock()
    mock_response.read.return_value = json.dumps(SAMPLE_RAW_RESPONSE).encode('utf-8')
    mock_urlopen.return_value = mock_response

    # Get agents
    agents = opamp_client.get_agents()

    # Verify request
    assert mock_urlopen.called
    request = mock_urlopen.call_args[0][0]
    assert request.full_url == 'https://opamp.example.com:4320/v1/agents'
    assert request.headers['Authorization'] == 'Bearer test-token'
    assert request.headers['Accept'] == 'application/json'

    # Verify response normalization
    assert len(agents) == 2

    # Check first agent
    agent1 = agents[0]
    assert agent1['name'] == 'web-01'
    assert agent1['id'].startswith('opamp://')
    assert agent1['tags']['rollout_group'] == 'production'
    assert agent1['tags']['region'] == 'us-east-1'
    assert agent1['properties']['agent_version'] == '0.96.0'
    assert agent1['properties']['health_status'] == 'healthy'
    assert agent1['properties']['service_name'] == 'otel-collector'
    assert agent1['provisioning_state'] == 'Connected'

    # Check second agent
    agent2 = agents[1]
    assert agent2['name'] == 'web-02'
    assert agent2['properties']['health_status'] == 'unhealthy'
    assert agent2['provisioning_state'] == 'Disconnected'


@patch('opamp_client.urlopen')
def test_get_agents_http_error(mock_urlopen, opamp_client):
    """Test handling of HTTP errors."""
    # Mock HTTP 401 error
    error_response = MagicMock()
    error_response.read.return_value = b'{"error": "unauthorized"}'
    mock_urlopen.side_effect = HTTPError(
        'https://opamp.example.com:4320/v1/agents',
        401,
        'Unauthorized',
        {},
        error_response
    )

    # Should raise HTTPError
    with pytest.raises(HTTPError) as exc_info:
        opamp_client.get_agents()

    assert 'OpAMP API error' in str(exc_info.value)


@patch('opamp_client.urlopen')
def test_get_agents_connection_error(mock_urlopen, opamp_client):
    """Test handling of connection errors."""
    # Mock connection error
    mock_urlopen.side_effect = URLError('Connection refused')

    # Should raise URLError
    with pytest.raises(URLError) as exc_info:
        opamp_client.get_agents()

    assert 'Failed to connect to OpAMP server' in str(exc_info.value)


@patch('opamp_client.urlopen')
def test_get_agents_invalid_json(mock_urlopen, opamp_client):
    """Test handling of invalid JSON response."""
    # Mock invalid JSON response
    mock_response = MagicMock()
    mock_response.read.return_value = b'not valid json'
    mock_urlopen.return_value = mock_response

    # Should raise ValueError
    with pytest.raises(ValueError) as exc_info:
        opamp_client.get_agents()

    assert 'Invalid JSON response' in str(exc_info.value)


@patch('opamp_client.urlopen')
def test_get_agents_non_list_response(mock_urlopen, opamp_client):
    """Test handling of non-list response."""
    # Mock response that's a dict instead of list
    mock_response = MagicMock()
    mock_response.read.return_value = json.dumps({'agents': []}).encode('utf-8')
    mock_urlopen.return_value = mock_response

    # Should raise ValueError
    with pytest.raises(ValueError) as exc_info:
        opamp_client.get_agents()

    assert 'Expected list of agents' in str(exc_info.value)


def test_normalize_agent(opamp_client):
    """Test agent normalization to Azure taxonomy."""
    raw_agent = SAMPLE_RAW_RESPONSE[0]
    normalized = opamp_client._normalize_agent(raw_agent)

    # Verify structure
    assert 'id' in normalized
    assert 'name' in normalized
    assert 'location' in normalized
    assert 'tags' in normalized
    assert 'properties' in normalized
    assert 'provisioning_state' in normalized

    # Verify values
    assert normalized['name'] == 'web-01'
    assert normalized['location'] == 'web-01'
    assert normalized['tags'] == {'rollout_group': 'production', 'region': 'us-east-1'}
    assert normalized['properties']['agent_version'] == '0.96.0'
    assert normalized['properties']['health_status'] == 'healthy'
    assert normalized['provisioning_state'] == 'Connected'


def test_filter_agents_by_name(opamp_client):
    """Test filtering agents by name."""
    agents = [
        {'name': 'web-01', 'tags': {}},
        {'name': 'web-02', 'tags': {}},
        {'name': 'db-01', 'tags': {}}
    ]

    filtered = opamp_client.filter_agents(agents, name='web-01')
    assert len(filtered) == 1
    assert filtered[0]['name'] == 'web-01'


def test_filter_agents_by_tags(opamp_client):
    """Test filtering agents by tags."""
    agents = [
        {'name': 'web-01', 'tags': {'rollout_group': 'production', 'region': 'us-east-1'}},
        {'name': 'web-02', 'tags': {'rollout_group': 'staging', 'region': 'us-west-2'}},
        {'name': 'web-03', 'tags': {'rollout_group': 'production', 'region': 'us-west-2'}}
    ]

    # Filter by single tag
    filtered = opamp_client.filter_agents(agents, tags={'rollout_group': 'production'})
    assert len(filtered) == 2
    assert all(a['tags']['rollout_group'] == 'production' for a in filtered)

    # Filter by multiple tags (AND logic)
    filtered = opamp_client.filter_agents(
        agents,
        tags={'rollout_group': 'production', 'region': 'us-east-1'}
    )
    assert len(filtered) == 1
    assert filtered[0]['name'] == 'web-01'


def test_filter_agents_by_name_and_tags(opamp_client):
    """Test filtering agents by both name and tags."""
    agents = [
        {'name': 'web-01', 'tags': {'rollout_group': 'production'}},
        {'name': 'web-02', 'tags': {'rollout_group': 'production'}},
        {'name': 'web-03', 'tags': {'rollout_group': 'staging'}}
    ]

    filtered = opamp_client.filter_agents(
        agents,
        name='web-01',
        tags={'rollout_group': 'production'}
    )
    assert len(filtered) == 1
    assert filtered[0]['name'] == 'web-01'


def test_filter_agents_no_match(opamp_client):
    """Test filtering with no matching agents."""
    agents = [
        {'name': 'web-01', 'tags': {'rollout_group': 'production'}},
        {'name': 'web-02', 'tags': {'rollout_group': 'staging'}}
    ]

    filtered = opamp_client.filter_agents(agents, name='nonexistent')
    assert len(filtered) == 0

    filtered = opamp_client.filter_agents(agents, tags={'rollout_group': 'development'})
    assert len(filtered) == 0
