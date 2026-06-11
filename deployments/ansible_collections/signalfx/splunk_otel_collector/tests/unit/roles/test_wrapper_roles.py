#!/usr/bin/env python3
# Copyright Splunk Inc.
# SPDX-License-Identifier: Apache-2.0

"""Unit tests for wrapper roles (host_monitoring and app_instrumentation)."""

import os
import pytest
import yaml


COLLECTION_ROOT = os.path.join(
    os.path.dirname(__file__), "..", "..", ".."
)


class TestHostMonitoringRole:
    """Test suite for host_monitoring role."""

    @pytest.fixture
    def defaults(self):
        """Load host_monitoring defaults."""
        defaults_path = os.path.join(
            COLLECTION_ROOT, "roles", "host_monitoring", "defaults", "main.yml"
        )
        with open(defaults_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    @pytest.fixture
    def tasks(self):
        """Load host_monitoring tasks."""
        tasks_path = os.path.join(
            COLLECTION_ROOT, "roles", "host_monitoring", "tasks", "main.yml"
        )
        with open(tasks_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    @pytest.fixture
    def meta(self):
        """Load host_monitoring meta."""
        meta_path = os.path.join(
            COLLECTION_ROOT, "roles", "host_monitoring", "meta", "main.yml"
        )
        with open(meta_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    def test_defaults_exist(self, defaults):
        """Test that defaults file is valid and contains expected variables."""
        assert defaults is not None
        assert "otel_host_monitoring_config_path" in defaults
        assert "otel_host_monitoring_gateway" in defaults
        assert "otel_host_monitoring_metrics" in defaults
        assert "otel_host_monitoring_log_paths" in defaults

    def test_gateway_default_empty(self, defaults):
        """Test that gateway default is empty (requires user input)."""
        assert defaults["otel_host_monitoring_gateway"] == ""

    def test_metrics_enabled_by_default(self, defaults):
        """Test that metrics collection is enabled by default."""
        assert defaults["otel_host_monitoring_metrics"] is True

    def test_log_paths_is_list(self, defaults):
        """Test that log_paths is a list."""
        assert isinstance(defaults["otel_host_monitoring_log_paths"], list)
        assert len(defaults["otel_host_monitoring_log_paths"]) > 0

    def test_memory_limits_are_numbers(self, defaults):
        """Test that memory limits are numeric."""
        assert isinstance(defaults["otel_host_monitoring_memory_limit_mib"], int)
        assert isinstance(defaults["otel_host_monitoring_memory_spike_limit_mib"], int)
        assert defaults["otel_host_monitoring_memory_limit_mib"] > 0
        assert defaults["otel_host_monitoring_memory_spike_limit_mib"] > 0

    def test_syslog_config_valid(self, defaults):
        """Test that syslog configuration is valid."""
        assert defaults["otel_host_monitoring_syslog_enabled"] is True
        assert defaults["otel_host_monitoring_syslog_protocol"] in ["rfc3164", "rfc5424"]
        assert ":" in defaults["otel_host_monitoring_syslog_listen"]

    def test_tasks_valid_yaml(self, tasks):
        """Test that tasks file is valid YAML."""
        assert isinstance(tasks, list)
        assert len(tasks) > 0

    def test_validation_task_exists(self, tasks):
        """Test that validation task exists and checks gateway."""
        validation_task = tasks[0]
        assert "ansible.builtin.assert" in validation_task
        assert "Validate required variables" in str(validation_task)

    def test_uses_pipeline_module(self, tasks):
        """Test that tasks use otel_collector_pipeline module."""
        task_str = yaml.dump(tasks)
        assert "signalfx.splunk_otel_collector.otel_collector_pipeline" in task_str

    def test_creates_metrics_pipeline(self, tasks):
        """Test that metrics pipeline is created."""
        task_str = yaml.dump(tasks)
        assert "metrics/host" in task_str
        assert "hostmetrics" in task_str

    def test_creates_logs_pipeline(self, tasks):
        """Test that logs pipeline is created."""
        task_str = yaml.dump(tasks)
        assert "logs/host" in task_str
        assert "filelog" in task_str

    def test_meta_has_license(self, meta):
        """Test that meta contains Apache-2.0 license."""
        assert meta["galaxy_info"]["license"] == "Apache-2.0"

    def test_meta_has_role_name(self, meta):
        """Test that meta contains correct role name."""
        assert meta["galaxy_info"]["role_name"] == "host_monitoring"

    def test_meta_has_min_ansible_version(self, meta):
        """Test that meta specifies minimum Ansible version."""
        assert "min_ansible_version" in meta["galaxy_info"]
        assert meta["galaxy_info"]["min_ansible_version"] == "2.14"


class TestAppInstrumentationRole:
    """Test suite for app_instrumentation role."""

    @pytest.fixture
    def defaults(self):
        """Load app_instrumentation defaults."""
        defaults_path = os.path.join(
            COLLECTION_ROOT, "roles", "app_instrumentation", "defaults", "main.yml"
        )
        with open(defaults_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    @pytest.fixture
    def tasks(self):
        """Load app_instrumentation tasks."""
        tasks_path = os.path.join(
            COLLECTION_ROOT, "roles", "app_instrumentation", "tasks", "main.yml"
        )
        with open(tasks_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    @pytest.fixture
    def meta(self):
        """Load app_instrumentation meta."""
        meta_path = os.path.join(
            COLLECTION_ROOT, "roles", "app_instrumentation", "meta", "main.yml"
        )
        with open(meta_path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f)

    def test_defaults_exist(self, defaults):
        """Test that defaults file is valid and contains expected variables."""
        assert defaults is not None
        assert "otel_app_config_path" in defaults
        assert "otel_app_gateway" in defaults
        assert "otel_app_signal_types" in defaults
        assert "otel_app_sampling_ratio" in defaults

    def test_gateway_default_empty(self, defaults):
        """Test that gateway default is empty (requires user input)."""
        assert defaults["otel_app_gateway"] == ""

    def test_signal_types_is_list(self, defaults):
        """Test that signal_types is a list with valid values."""
        assert isinstance(defaults["otel_app_signal_types"], list)
        assert len(defaults["otel_app_signal_types"]) > 0
        valid_signals = ["traces", "metrics", "logs"]
        for signal in defaults["otel_app_signal_types"]:
            assert signal in valid_signals

    def test_sampling_ratio_valid(self, defaults):
        """Test that sampling_ratio is a valid float."""
        assert isinstance(defaults["otel_app_sampling_ratio"], (int, float))
        assert 0.0 <= defaults["otel_app_sampling_ratio"] <= 1.0

    def test_listen_addresses_valid(self, defaults):
        """Test that listen addresses are valid."""
        assert ":" in defaults["otel_app_listen_grpc"]
        assert ":" in defaults["otel_app_listen_http"]
        assert "4317" in defaults["otel_app_listen_grpc"]
        assert "4318" in defaults["otel_app_listen_http"]

    def test_memory_limits_are_numbers(self, defaults):
        """Test that memory limits are numeric."""
        assert isinstance(defaults["otel_app_memory_limit_mib"], int)
        assert isinstance(defaults["otel_app_memory_spike_limit_mib"], int)
        assert defaults["otel_app_memory_limit_mib"] > 0
        assert defaults["otel_app_memory_spike_limit_mib"] > 0

    def test_tasks_valid_yaml(self, tasks):
        """Test that tasks file is valid YAML."""
        assert isinstance(tasks, list)
        assert len(tasks) > 0

    def test_validation_task_exists(self, tasks):
        """Test that validation task exists and checks gateway."""
        validation_task = tasks[0]
        assert "ansible.builtin.assert" in validation_task
        assert "Validate required variables" in str(validation_task)

    def test_uses_pipeline_module(self, tasks):
        """Test that tasks use otel_collector_pipeline module."""
        task_str = yaml.dump(tasks)
        assert "signalfx.splunk_otel_collector.otel_collector_pipeline" in task_str

    def test_creates_pipelines_for_signals(self, tasks):
        """Test that pipelines are created for each signal type."""
        task_str = yaml.dump(tasks)
        assert "loop:" in task_str or "with_items:" in task_str
        assert "/app" in task_str
        assert "otlp" in task_str

    def test_otlp_receiver_configured(self, tasks):
        """Test that OTLP receiver is configured with gRPC and HTTP."""
        task_str = yaml.dump(tasks)
        assert "otlp" in task_str
        assert "grpc" in task_str.lower()
        assert "http" in task_str.lower()

    def test_sampling_processor_conditional(self, tasks):
        """Test that sampling processor is conditionally included."""
        task_str = yaml.dump(tasks)
        assert "probabilistic_sampler" in task_str
        assert "otel_app_sampling_ratio" in task_str

    def test_meta_has_license(self, meta):
        """Test that meta contains Apache-2.0 license."""
        assert meta["galaxy_info"]["license"] == "Apache-2.0"

    def test_meta_has_role_name(self, meta):
        """Test that meta contains correct role name."""
        assert meta["galaxy_info"]["role_name"] == "app_instrumentation"

    def test_meta_has_min_ansible_version(self, meta):
        """Test that meta specifies minimum Ansible version."""
        assert "min_ansible_version" in meta["galaxy_info"]
        assert meta["galaxy_info"]["min_ansible_version"] == "2.14"


class TestRoleStructure:
    """Test that roles have proper structure."""

    def test_host_monitoring_structure(self):
        """Test that host_monitoring role has required files."""
        role_path = os.path.join(COLLECTION_ROOT, "roles", "host_monitoring")
        assert os.path.exists(os.path.join(role_path, "defaults", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "tasks", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "meta", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "README.md"))

    def test_app_instrumentation_structure(self):
        """Test that app_instrumentation role has required files."""
        role_path = os.path.join(COLLECTION_ROOT, "roles", "app_instrumentation")
        assert os.path.exists(os.path.join(role_path, "defaults", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "tasks", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "meta", "main.yml"))
        assert os.path.exists(os.path.join(role_path, "README.md"))

    def test_host_monitoring_readme_content(self):
        """Test that host_monitoring README contains key sections."""
        readme_path = os.path.join(
            COLLECTION_ROOT, "roles", "host_monitoring", "README.md"
        )
        with open(readme_path, "r", encoding="utf-8") as f:
            content = f.read()
        assert "# host_monitoring" in content
        assert "Requirements" in content
        assert "Role Variables" in content
        assert "Example Playbook" in content
        assert "License" in content
        assert "Apache-2.0" in content

    def test_app_instrumentation_readme_content(self):
        """Test that app_instrumentation README contains key sections."""
        readme_path = os.path.join(
            COLLECTION_ROOT, "roles", "app_instrumentation", "README.md"
        )
        with open(readme_path, "r", encoding="utf-8") as f:
            content = f.read()
        assert "# app_instrumentation" in content
        assert "Requirements" in content
        assert "Role Variables" in content
        assert "Example Playbook" in content
        assert "License" in content
        assert "Apache-2.0" in content
