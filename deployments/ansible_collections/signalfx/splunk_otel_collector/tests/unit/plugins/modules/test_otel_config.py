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

"""Tests for OtelConfig utility."""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import os
import pytest
from pathlib import Path

try:
    from ruamel.yaml import YAML
    HAS_RUAMEL_YAML = True
except ImportError:
    HAS_RUAMEL_YAML = False

# Import the module under test
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))
from otel_config import OtelConfig


SAMPLE_CONFIG = """
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
  syslog:
    protocol: rfc5424

processors:
  batch:
    timeout: 5s
  memory_limiter:
    limit_mib: 512

exporters:
  otlphttp:
    endpoint: "https://gateway:4318"

service:
  pipelines:
    logs:
      receivers: [syslog]
      processors: [batch, memory_limiter]
      exporters: [otlphttp]
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp]
"""


@pytest.fixture
def config_file(tmp_path):
    """Create a temporary config file."""
    config_path = tmp_path / "agent_config.yaml"
    config_path.write_text(SAMPLE_CONFIG)
    return str(config_path)


@pytest.fixture
def empty_config_file(tmp_path):
    """Return path to non-existent config file."""
    return str(tmp_path / "empty_config.yaml")


@pytest.mark.skipif(not HAS_RUAMEL_YAML, reason="ruamel.yaml not available")
class TestOtelConfig:
    """Test OtelConfig utility class."""

    def test_load_existing_config(self, config_file):
        """Test loading an existing config file."""
        config = OtelConfig(config_file)
        data = config.load()

        assert data is not None
        assert 'receivers' in data
        assert 'processors' in data
        assert 'exporters' in data
        assert 'service' in data
        assert 'pipelines' in data['service']

    def test_load_nonexistent_config(self, empty_config_file):
        """Test loading a non-existent config file returns empty structure."""
        config = OtelConfig(empty_config_file)
        data = config.load()

        assert data is not None
        assert data == {
            'receivers': {},
            'processors': {},
            'exporters': {},
            'service': {'pipelines': {}}
        }

    def test_save_config(self, config_file):
        """Test saving config to file."""
        config = OtelConfig(config_file)
        config.load()

        # Modify and save
        config.set_component('receivers', 'test_receiver', {'endpoint': 'localhost:1234'})
        config.save()

        # Reload and verify
        config2 = OtelConfig(config_file)
        data = config2.load()
        assert 'test_receiver' in data['receivers']
        assert data['receivers']['test_receiver']['endpoint'] == 'localhost:1234'

    def test_save_with_backup(self, config_file):
        """Test saving with backup creates backup file."""
        config = OtelConfig(config_file)
        config.load()

        # Modify and save with backup
        config.set_component('receivers', 'test_receiver', {'endpoint': 'localhost:1234'})
        config.save(backup=True)

        # Check backup file exists
        config_dir = Path(config_file).parent
        backup_files = list(config_dir.glob("*.bak"))
        assert len(backup_files) == 1

    def test_get_component(self, config_file):
        """Test getting a component."""
        config = OtelConfig(config_file)
        config.load()

        receiver = config.get_component('receivers', 'otlp')
        assert receiver is not None
        assert 'protocols' in receiver

        processor = config.get_component('processors', 'batch')
        assert processor is not None
        assert processor['timeout'] == '5s'

        exporter = config.get_component('exporters', 'otlphttp')
        assert exporter is not None
        assert exporter['endpoint'] == 'https://gateway:4318'

    def test_get_nonexistent_component(self, config_file):
        """Test getting a non-existent component returns None."""
        config = OtelConfig(config_file)
        config.load()

        result = config.get_component('receivers', 'nonexistent')
        assert result is None

    def test_set_component(self, config_file):
        """Test setting a component."""
        config = OtelConfig(config_file)
        config.load()

        config.set_component('receivers', 'new_receiver', {
            'endpoint': 'localhost:5555'
        })

        receiver = config.get_component('receivers', 'new_receiver')
        assert receiver is not None
        assert receiver['endpoint'] == 'localhost:5555'

    def test_update_existing_component(self, config_file):
        """Test updating an existing component."""
        config = OtelConfig(config_file)
        config.load()

        config.set_component('processors', 'batch', {
            'timeout': '10s',
            'send_batch_size': 100
        })

        processor = config.get_component('processors', 'batch')
        assert processor['timeout'] == '10s'
        assert processor['send_batch_size'] == 100

    def test_remove_component(self, config_file):
        """Test removing a component."""
        config = OtelConfig(config_file)
        config.load()

        # Component exists
        assert config.get_component('receivers', 'syslog') is not None

        # Remove it
        result = config.remove_component('receivers', 'syslog')
        assert result is True

        # Component is gone
        assert config.get_component('receivers', 'syslog') is None

    def test_remove_component_from_pipelines(self, config_file):
        """Test removing a component also removes it from pipeline references."""
        config = OtelConfig(config_file)
        config.load()

        # Verify syslog is in logs pipeline
        logs_pipeline = config.get_pipeline('logs')
        assert 'syslog' in logs_pipeline['receivers']

        # Remove syslog receiver
        config.remove_component('receivers', 'syslog')

        # Verify it's removed from pipeline
        logs_pipeline = config.get_pipeline('logs')
        assert 'syslog' not in logs_pipeline['receivers']

    def test_remove_nonexistent_component(self, config_file):
        """Test removing a non-existent component returns False."""
        config = OtelConfig(config_file)
        config.load()

        result = config.remove_component('receivers', 'nonexistent')
        assert result is False

    def test_get_pipeline(self, config_file):
        """Test getting a pipeline."""
        config = OtelConfig(config_file)
        config.load()

        pipeline = config.get_pipeline('logs')
        assert pipeline is not None
        assert 'receivers' in pipeline
        assert 'processors' in pipeline
        assert 'exporters' in pipeline
        assert pipeline['receivers'] == ['syslog']
        assert pipeline['processors'] == ['batch', 'memory_limiter']
        assert pipeline['exporters'] == ['otlphttp']

    def test_get_nonexistent_pipeline(self, config_file):
        """Test getting a non-existent pipeline returns None."""
        config = OtelConfig(config_file)
        config.load()

        result = config.get_pipeline('nonexistent')
        assert result is None

    def test_set_pipeline(self, config_file):
        """Test setting a pipeline."""
        config = OtelConfig(config_file)
        config.load()

        config.set_pipeline(
            'metrics',
            receivers=['prometheus'],
            processors=['batch'],
            exporters=['otlphttp']
        )

        pipeline = config.get_pipeline('metrics')
        assert pipeline is not None
        assert pipeline['receivers'] == ['prometheus']
        assert pipeline['processors'] == ['batch']
        assert pipeline['exporters'] == ['otlphttp']

    def test_update_existing_pipeline(self, config_file):
        """Test updating an existing pipeline."""
        config = OtelConfig(config_file)
        config.load()

        config.set_pipeline(
            'logs',
            receivers=['syslog', 'filelog'],
            processors=['batch'],
            exporters=['otlphttp', 'logging']
        )

        pipeline = config.get_pipeline('logs')
        assert pipeline['receivers'] == ['syslog', 'filelog']
        assert pipeline['processors'] == ['batch']
        assert pipeline['exporters'] == ['otlphttp', 'logging']

    def test_remove_pipeline(self, config_file):
        """Test removing a pipeline."""
        config = OtelConfig(config_file)
        config.load()

        # Pipeline exists
        assert config.get_pipeline('logs') is not None

        # Remove it
        result = config.remove_pipeline('logs')
        assert result is True

        # Pipeline is gone
        assert config.get_pipeline('logs') is None

    def test_remove_nonexistent_pipeline(self, config_file):
        """Test removing a non-existent pipeline returns False."""
        config = OtelConfig(config_file)
        config.load()

        result = config.remove_pipeline('nonexistent')
        assert result is False

    def test_has_changed_no_changes(self, config_file):
        """Test has_changed returns False when no changes made."""
        config = OtelConfig(config_file)
        config.load()

        assert config.has_changed() is False

    def test_has_changed_after_modification(self, config_file):
        """Test has_changed returns True after modification."""
        config = OtelConfig(config_file)
        config.load()

        config.set_component('receivers', 'new_receiver', {'endpoint': 'localhost:1234'})

        assert config.has_changed() is True

    def test_has_changed_after_save(self, config_file):
        """Test has_changed returns False after saving changes."""
        config = OtelConfig(config_file)
        config.load()

        config.set_component('receivers', 'new_receiver', {'endpoint': 'localhost:1234'})
        assert config.has_changed() is True

        config.save()
        assert config.has_changed() is False

    def test_create_new_config_file(self, empty_config_file):
        """Test creating a new config file from scratch."""
        config = OtelConfig(empty_config_file)
        config.load()

        config.set_component('receivers', 'otlp', {
            'protocols': {
                'grpc': {
                    'endpoint': '0.0.0.0:4317'
                }
            }
        })
        config.set_component('exporters', 'otlphttp', {
            'endpoint': 'https://gateway:4318'
        })
        config.set_pipeline(
            'traces',
            receivers=['otlp'],
            processors=[],
            exporters=['otlphttp']
        )

        config.save()

        # Verify file was created and can be loaded
        assert Path(empty_config_file).exists()

        config2 = OtelConfig(empty_config_file)
        data = config2.load()

        assert 'otlp' in data['receivers']
        assert 'otlphttp' in data['exporters']
        assert 'traces' in data['service']['pipelines']
