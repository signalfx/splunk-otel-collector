# Copyright Splunk Inc.
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

"""Tests for the gateway role templates."""

import os
import pytest
import yaml
from jinja2 import Environment, FileSystemLoader


@pytest.fixture
def template_dir():
    """Get the path to the gateway role templates directory."""
    current_dir = os.path.dirname(os.path.abspath(__file__))
    collection_root = os.path.join(current_dir, "..", "..", "..")
    template_path = os.path.join(
        collection_root, "roles", "gateway", "templates"
    )
    return os.path.abspath(template_path)


@pytest.fixture
def jinja_env(template_dir):
    """Create a Jinja2 environment for rendering templates."""
    return Environment(loader=FileSystemLoader(template_dir))


@pytest.fixture
def sample_vars():
    """Sample variables for template rendering."""
    return {
        "otel_gateway_listen_host": "0.0.0.0",
        "otel_gateway_otlp_http_port": 4318,
        "otel_gateway_otlp_grpc_port": 4317,
        "otel_gateway_health_port": 13133,
        "otel_gateway_memory_limit_mib": 2048,
        "otel_gateway_memory_spike_limit_mib": 512,
        "otel_gateway_batch_timeout": "5s",
        "otel_gateway_batch_send_size": 8192,
        "otel_gateway_tls_enabled": False,
        "otel_gateway_exporters": [
            {
                "name": "otlphttp",
                "endpoint": "https://backend:4318"
            },
            {
                "name": "splunk_hec",
                "endpoint": "https://splunk:8088/services/collector",
                "token": "secret-token-123"
            }
        ]
    }


@pytest.fixture
def tls_vars(sample_vars):
    """Sample variables with TLS enabled."""
    tls_config = sample_vars.copy()
    tls_config["otel_gateway_tls_enabled"] = True
    tls_config["otel_gateway_tls_cert"] = "/etc/ssl/certs/gateway.crt"
    tls_config["otel_gateway_tls_key"] = "/etc/ssl/private/gateway.key"
    tls_config["otel_gateway_tls_ca"] = "/etc/ssl/certs/ca.crt"
    return tls_config


def test_gateway_config_template_exists(template_dir):
    """Test that the gateway config template exists."""
    template_file = os.path.join(template_dir, "gateway_config.yaml.j2")
    assert os.path.exists(template_file), "gateway_config.yaml.j2 not found"


def test_gateway_config_renders(jinja_env, sample_vars):
    """Test that the gateway config template renders without errors."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    assert rendered is not None
    assert len(rendered) > 0


def test_gateway_config_is_valid_yaml(jinja_env, sample_vars):
    """Test that the rendered gateway config is valid YAML."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)
    assert config is not None
    assert isinstance(config, dict)


def test_gateway_config_has_required_sections(jinja_env, sample_vars):
    """Test that the gateway config has all required sections."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    # Check required top-level sections
    assert "receivers" in config
    assert "processors" in config
    assert "exporters" in config
    assert "service" in config
    assert "extensions" in config


def test_gateway_config_has_otlp_receiver(jinja_env, sample_vars):
    """Test that the gateway config has OTLP receiver configured."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    assert "otlp" in config["receivers"]
    assert "protocols" in config["receivers"]["otlp"]
    assert "grpc" in config["receivers"]["otlp"]["protocols"]
    assert "http" in config["receivers"]["otlp"]["protocols"]

    # Check endpoints
    grpc = config["receivers"]["otlp"]["protocols"]["grpc"]
    assert grpc["endpoint"] == "0.0.0.0:4317"

    http = config["receivers"]["otlp"]["protocols"]["http"]
    assert http["endpoint"] == "0.0.0.0:4318"


def test_gateway_config_has_processors(jinja_env, sample_vars):
    """Test that the gateway config has required processors."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    assert "memory_limiter" in config["processors"]
    assert "batch" in config["processors"]

    # Check memory_limiter config
    memory_limiter = config["processors"]["memory_limiter"]
    assert memory_limiter["limit_mib"] == 2048
    assert memory_limiter["spike_limit_mib"] == 512

    # Check batch config
    batch = config["processors"]["batch"]
    assert batch["timeout"] == "5s"
    assert batch["send_batch_size"] == 8192


def test_gateway_config_has_exporters(jinja_env, sample_vars):
    """Test that the gateway config has exporters configured."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    assert "otlphttp" in config["exporters"]
    assert "splunk_hec" in config["exporters"]

    # Check exporter endpoints
    assert config["exporters"]["otlphttp"]["endpoint"] == "https://backend:4318"
    assert config["exporters"]["splunk_hec"]["endpoint"] == "https://splunk:8088/services/collector"
    assert config["exporters"]["splunk_hec"]["token"] == "secret-token-123"


def test_gateway_config_has_service_pipelines(jinja_env, sample_vars):
    """Test that the gateway config has service pipelines configured."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    assert "pipelines" in config["service"]
    pipelines = config["service"]["pipelines"]

    # Check all three pipeline types
    assert "traces" in pipelines
    assert "metrics" in pipelines
    assert "logs" in pipelines

    # Check traces pipeline
    traces = pipelines["traces"]
    assert traces["receivers"] == ["otlp"]
    assert "memory_limiter" in traces["processors"]
    assert "batch" in traces["processors"]
    assert "otlphttp" in traces["exporters"]
    assert "splunk_hec" in traces["exporters"]

    # Check metrics pipeline
    metrics = pipelines["metrics"]
    assert metrics["receivers"] == ["otlp"]
    assert "memory_limiter" in metrics["processors"]
    assert "batch" in metrics["processors"]

    # Check logs pipeline
    logs = pipelines["logs"]
    assert logs["receivers"] == ["otlp"]
    assert "memory_limiter" in logs["processors"]
    assert "batch" in logs["processors"]


def test_gateway_config_has_extensions(jinja_env, sample_vars):
    """Test that the gateway config has extensions configured."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    assert "health_check" in config["extensions"]
    assert config["extensions"]["health_check"]["endpoint"] == "0.0.0.0:13133"

    # Check service extensions
    assert "extensions" in config["service"]
    assert "health_check" in config["service"]["extensions"]


def test_gateway_config_with_tls(jinja_env, tls_vars):
    """Test that the gateway config includes TLS when enabled."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(tls_vars)
    config = yaml.safe_load(rendered)

    # Check gRPC TLS
    grpc = config["receivers"]["otlp"]["protocols"]["grpc"]
    assert "tls" in grpc
    assert grpc["tls"]["cert_file"] == "/etc/ssl/certs/gateway.crt"
    assert grpc["tls"]["key_file"] == "/etc/ssl/private/gateway.key"
    assert grpc["tls"]["ca_file"] == "/etc/ssl/certs/ca.crt"

    # Check HTTP TLS
    http = config["receivers"]["otlp"]["protocols"]["http"]
    assert "tls" in http
    assert http["tls"]["cert_file"] == "/etc/ssl/certs/gateway.crt"
    assert http["tls"]["key_file"] == "/etc/ssl/private/gateway.key"
    assert http["tls"]["ca_file"] == "/etc/ssl/certs/ca.crt"


def test_gateway_config_without_tls(jinja_env, sample_vars):
    """Test that the gateway config excludes TLS when disabled."""
    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(sample_vars)
    config = yaml.safe_load(rendered)

    # Check gRPC has no TLS
    grpc = config["receivers"]["otlp"]["protocols"]["grpc"]
    assert "tls" not in grpc

    # Check HTTP has no TLS
    http = config["receivers"]["otlp"]["protocols"]["http"]
    assert "tls" not in http


def test_gateway_env_template_renders(jinja_env, sample_vars):
    """Test that the gateway environment template renders correctly."""
    env_vars = sample_vars.copy()
    env_vars["otel_gateway_config_path"] = "/etc/otel/collector/gateway_config.yaml"
    env_vars["otel_gateway_additional_env_vars"] = {
        "DEBUG": "true",
        "LOG_LEVEL": "info"
    }

    template = jinja_env.get_template("gateway.conf.j2")
    rendered = template.render(env_vars)

    assert "SPLUNK_CONFIG=/etc/otel/collector/gateway_config.yaml" in rendered
    assert "SPLUNK_MEMORY_TOTAL_MIB=2048" in rendered
    assert "OTEL_GATEWAY_LISTEN_HOST=0.0.0.0" in rendered
    assert "OTEL_GATEWAY_OTLP_HTTP_PORT=4318" in rendered
    assert "OTEL_GATEWAY_OTLP_GRPC_PORT=4317" in rendered
    assert "DEBUG=true" in rendered
    assert "LOG_LEVEL=info" in rendered


def test_gateway_service_template_renders(jinja_env):
    """Test that the gateway service template renders correctly."""
    service_vars = {
        "otel_gateway_env_file": "/etc/otel/collector/gateway.conf",
        "otel_gateway_user": "otel-collector",
        "otel_gateway_group": "otel-collector"
    }

    template = jinja_env.get_template("gateway.service.j2")
    rendered = template.render(service_vars)

    assert "EnvironmentFile=/etc/otel/collector/gateway.conf" in rendered
    assert "User=otel-collector" in rendered
    assert "Group=otel-collector" in rendered
    assert "ExecStart=/usr/bin/otelcol --config=${SPLUNK_CONFIG}" in rendered
    assert "Restart=on-failure" in rendered


def test_exporter_with_headers(jinja_env, sample_vars):
    """Test that exporters with headers are configured correctly."""
    vars_with_headers = sample_vars.copy()
    vars_with_headers["otel_gateway_exporters"] = [
        {
            "name": "otlphttp",
            "endpoint": "https://backend:4318",
            "headers": {
                "X-Auth-Token": "token123",
                "X-Tenant-ID": "tenant456"
            }
        }
    ]

    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(vars_with_headers)
    config = yaml.safe_load(rendered)

    exporter = config["exporters"]["otlphttp"]
    assert "headers" in exporter
    assert exporter["headers"]["X-Auth-Token"] == "token123"
    assert exporter["headers"]["X-Tenant-ID"] == "tenant456"


def test_exporter_with_timeout(jinja_env, sample_vars):
    """Test that exporters with timeout are configured correctly."""
    vars_with_timeout = sample_vars.copy()
    vars_with_timeout["otel_gateway_exporters"] = [
        {
            "name": "otlphttp",
            "endpoint": "https://backend:4318",
            "timeout": "30s"
        }
    ]

    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(vars_with_timeout)
    config = yaml.safe_load(rendered)

    exporter = config["exporters"]["otlphttp"]
    assert exporter["timeout"] == "30s"


def test_multiple_exporters_in_pipeline(jinja_env, sample_vars):
    """Test that multiple exporters are correctly added to pipelines."""
    vars_multi = sample_vars.copy()
    vars_multi["otel_gateway_exporters"] = [
        {"name": "otlphttp/primary", "endpoint": "https://primary:4318"},
        {"name": "otlphttp/secondary", "endpoint": "https://secondary:4318"},
        {"name": "otlphttp/tertiary", "endpoint": "https://tertiary:4318"}
    ]

    template = jinja_env.get_template("gateway_config.yaml.j2")
    rendered = template.render(vars_multi)
    config = yaml.safe_load(rendered)

    # Check all exporters are in the traces pipeline
    traces_exporters = config["service"]["pipelines"]["traces"]["exporters"]
    assert "otlphttp/primary" in traces_exporters
    assert "otlphttp/secondary" in traces_exporters
    assert "otlphttp/tertiary" in traces_exporters
    assert len(traces_exporters) == 3
