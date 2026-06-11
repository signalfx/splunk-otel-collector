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

"""Tests for info modules."""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import os
import sys
import pytest
from pathlib import Path

# Add module paths
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/modules'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))


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

exporters:
  otlphttp:
    endpoint: "https://gateway:4318"

service:
  pipelines:
    logs:
      receivers: [syslog]
      processors: [batch]
      exporters: [otlphttp]
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp]
"""


class TestOtelCollectorInfo:
    """Test otel_collector_info module."""

    def test_module_imports(self):
        """Test that the module can be imported."""
        import otel_collector_info
        assert hasattr(otel_collector_info, 'main')
        assert hasattr(otel_collector_info, 'get_collector_info')

    def test_module_documentation(self):
        """Test that module has proper documentation."""
        import otel_collector_info
        assert hasattr(otel_collector_info, 'DOCUMENTATION')
        assert hasattr(otel_collector_info, 'EXAMPLES')
        assert hasattr(otel_collector_info, 'RETURN')

        # Verify key fields in documentation
        assert 'module: otel_collector_info' in otel_collector_info.DOCUMENTATION
        assert 'config_path:' in otel_collector_info.DOCUMENTATION
        assert 'health_endpoint:' in otel_collector_info.DOCUMENTATION

    def test_module_examples(self):
        """Test that module has examples."""
        import otel_collector_info
        assert 'otel_collector_info:' in otel_collector_info.EXAMPLES

    def test_module_returns(self):
        """Test that module documents return values."""
        import otel_collector_info
        assert 'collectors:' in otel_collector_info.RETURN
        assert 'provisioning_state:' in otel_collector_info.RETURN

    def test_get_machine_id(self):
        """Test machine ID retrieval."""
        import otel_collector_info
        machine_id = otel_collector_info.get_machine_id()
        assert machine_id is not None
        # Should return "unknown" or an actual ID
        assert isinstance(machine_id, str)

    def test_check_health_unknown(self):
        """Test health check with invalid endpoint."""
        import otel_collector_info
        status = otel_collector_info.check_health('http://invalid-endpoint:99999')
        assert status in ['Running', 'Stopped', 'Unknown']

    def test_infer_signal_type(self):
        """Test signal type inference."""
        import otel_pipeline_info
        assert otel_pipeline_info.infer_signal_type('logs') == 'logs'
        assert otel_pipeline_info.infer_signal_type('metrics') == 'metrics'
        assert otel_pipeline_info.infer_signal_type('traces') == 'traces'
        assert otel_pipeline_info.infer_signal_type('my-custom-logs-pipeline') == 'logs'
        assert otel_pipeline_info.infer_signal_type('custom') == 'unknown'


class TestOtelPipelineInfo:
    """Test otel_pipeline_info module."""

    def test_module_imports(self):
        """Test that the module can be imported."""
        import otel_pipeline_info
        assert hasattr(otel_pipeline_info, 'main')
        assert hasattr(otel_pipeline_info, 'get_pipeline_info')

    def test_module_documentation(self):
        """Test that module has proper documentation."""
        import otel_pipeline_info
        assert hasattr(otel_pipeline_info, 'DOCUMENTATION')
        assert hasattr(otel_pipeline_info, 'EXAMPLES')
        assert hasattr(otel_pipeline_info, 'RETURN')

        # Verify key fields in documentation
        assert 'module: otel_pipeline_info' in otel_pipeline_info.DOCUMENTATION
        assert 'config_path:' in otel_pipeline_info.DOCUMENTATION

    def test_module_examples(self):
        """Test that module has examples."""
        import otel_pipeline_info
        assert 'otel_pipeline_info:' in otel_pipeline_info.EXAMPLES

    def test_module_returns(self):
        """Test that module documents return values."""
        import otel_pipeline_info
        assert 'pipelines:' in otel_pipeline_info.RETURN
        assert 'signal_type:' in otel_pipeline_info.RETURN


class TestOtelReceiverInfo:
    """Test otel_receiver_info module."""

    def test_module_imports(self):
        """Test that the module can be imported."""
        import otel_receiver_info
        assert hasattr(otel_receiver_info, 'main')
        assert hasattr(otel_receiver_info, 'get_receiver_info')

    def test_module_documentation(self):
        """Test that module has proper documentation."""
        import otel_receiver_info
        assert hasattr(otel_receiver_info, 'DOCUMENTATION')
        assert hasattr(otel_receiver_info, 'EXAMPLES')
        assert hasattr(otel_receiver_info, 'RETURN')

        # Verify key fields in documentation
        assert 'module: otel_receiver_info' in otel_receiver_info.DOCUMENTATION
        assert 'config_path:' in otel_receiver_info.DOCUMENTATION

    def test_module_examples(self):
        """Test that module has examples."""
        import otel_receiver_info
        assert 'otel_receiver_info:' in otel_receiver_info.EXAMPLES

    def test_module_returns(self):
        """Test that module documents return values."""
        import otel_receiver_info
        assert 'receivers:' in otel_receiver_info.RETURN
        assert 'used_in_pipelines:' in otel_receiver_info.RETURN


class TestIntegrationWithOtelConfig:
    """Integration tests using OtelConfig."""

    def test_collector_info_with_config(self, tmp_path):
        """Test collector info using actual config file."""
        from unittest.mock import MagicMock, patch
        import otel_collector_info

        # Create config file
        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text(SAMPLE_CONFIG)

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'health_endpoint': 'http://localhost:13133',
            'name': None
        }

        with patch('otel_collector_info.socket.gethostname', return_value='test-host'), \
             patch('otel_collector_info.get_machine_id', return_value='abc123'), \
             patch('otel_collector_info.check_health', return_value='Running'):

            result = otel_collector_info.get_collector_info(mock_module)

            assert not result['changed']
            assert len(result['collectors']) == 1
            collector = result['collectors'][0]
            assert collector['location'] == 'test-host'
            assert 'logs' in collector['properties']['active_pipelines']
            assert 'traces' in collector['properties']['active_pipelines']

    def test_pipeline_info_with_config(self, tmp_path):
        """Test pipeline info using actual config file."""
        from unittest.mock import MagicMock, patch
        import otel_pipeline_info

        # Create config file
        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text(SAMPLE_CONFIG)

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'name': None
        }

        with patch('otel_pipeline_info.socket.gethostname', return_value='test-host'):
            result = otel_pipeline_info.get_pipeline_info(mock_module)

            assert not result['changed']
            assert len(result['pipelines']) == 2

            # Verify logs pipeline
            logs = next(p for p in result['pipelines'] if p['name'] == 'logs')
            assert logs['properties']['signal_type'] == 'logs'
            assert logs['properties']['receivers'] == ['syslog']
            assert logs['properties']['processors'] == ['batch']
            assert logs['properties']['exporters'] == ['otlphttp']

            # Verify traces pipeline
            traces = next(p for p in result['pipelines'] if p['name'] == 'traces')
            assert traces['properties']['signal_type'] == 'traces'
            assert traces['properties']['receivers'] == ['otlp']

    def test_receiver_info_with_config(self, tmp_path):
        """Test receiver info using actual config file."""
        from unittest.mock import MagicMock, patch
        import otel_receiver_info

        # Create config file
        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text(SAMPLE_CONFIG)

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'name': None
        }

        with patch('otel_receiver_info.socket.gethostname', return_value='test-host'):
            result = otel_receiver_info.get_receiver_info(mock_module)

            assert not result['changed']
            assert len(result['receivers']) == 2

            # Verify OTLP receiver
            otlp = next(r for r in result['receivers'] if r['name'] == 'otlp')
            assert 'protocols' in otlp['properties']['config']
            assert 'traces' in otlp['properties']['used_in_pipelines']

            # Verify syslog receiver
            syslog = next(r for r in result['receivers'] if r['name'] == 'syslog')
            assert syslog['properties']['config']['protocol'] == 'rfc5424'
            assert 'logs' in syslog['properties']['used_in_pipelines']

    def test_pipeline_info_filtered(self, tmp_path):
        """Test pipeline info with name filter."""
        from unittest.mock import MagicMock, patch
        import otel_pipeline_info

        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text(SAMPLE_CONFIG)

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'name': 'logs'
        }

        with patch('otel_pipeline_info.socket.gethostname', return_value='test-host'):
            result = otel_pipeline_info.get_pipeline_info(mock_module)

            assert len(result['pipelines']) == 1
            assert result['pipelines'][0]['name'] == 'logs'

    def test_receiver_info_filtered(self, tmp_path):
        """Test receiver info with name filter."""
        from unittest.mock import MagicMock, patch
        import otel_receiver_info

        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text(SAMPLE_CONFIG)

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'name': 'otlp'
        }

        with patch('otel_receiver_info.socket.gethostname', return_value='test-host'):
            result = otel_receiver_info.get_receiver_info(mock_module)

            assert len(result['receivers']) == 1
            assert result['receivers'][0]['name'] == 'otlp'

    def test_pipeline_info_empty_config(self, tmp_path):
        """Test pipeline info with empty config."""
        from unittest.mock import MagicMock, patch
        import otel_pipeline_info

        config_path = tmp_path / "agent_config.yaml"
        config_path.write_text("receivers: {}\nexporters: {}\nprocessors: {}\nservice:\n  pipelines: {}")

        mock_module = MagicMock()
        mock_module.params = {
            'config_path': str(config_path),
            'name': None
        }

        with patch('otel_pipeline_info.socket.gethostname', return_value='test-host'):
            result = otel_pipeline_info.get_pipeline_info(mock_module)

            assert len(result['pipelines']) == 0
