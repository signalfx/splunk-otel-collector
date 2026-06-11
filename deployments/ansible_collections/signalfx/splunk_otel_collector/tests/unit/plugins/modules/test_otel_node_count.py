#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

"""Tests for otel_node_count module.

These are basic integration-style tests that validate the module can be imported
and has the correct structure.
"""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import os
import sys

import pytest

# Add module paths
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/modules'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))


def test_module_imports():
    """Test that the module can be imported."""
    import otel_node_count
    assert hasattr(otel_node_count, 'run_module')
    assert hasattr(otel_node_count, 'main')


def test_module_documentation():
    """Test that module has proper documentation."""
    import otel_node_count
    assert hasattr(otel_node_count, 'DOCUMENTATION')
    assert hasattr(otel_node_count, 'EXAMPLES')
    assert hasattr(otel_node_count, 'RETURN')


def test_documentation_format():
    """Test that module documentation follows expected format."""
    import otel_node_count

    # Check that DOCUMENTATION contains key fields
    assert 'module: otel_node_count' in otel_node_count.DOCUMENTATION
    assert 'short_description:' in otel_node_count.DOCUMENTATION
    assert 'version_added:' in otel_node_count.DOCUMENTATION
    assert 'options:' in otel_node_count.DOCUMENTATION

    # Check that required options are documented
    assert 'sources:' in otel_node_count.DOCUMENTATION
    assert 'validate_certs:' in otel_node_count.DOCUMENTATION
    assert 'timeout:' in otel_node_count.DOCUMENTATION


def test_examples_format():
    """Test that EXAMPLES section contains valid YAML-like content."""
    import otel_node_count

    assert 'signalfx.splunk_otel_collector.otel_node_count' in otel_node_count.EXAMPLES
    assert 'sources:' in otel_node_count.EXAMPLES
    assert 'type: opamp' in otel_node_count.EXAMPLES or 'type:' in otel_node_count.EXAMPLES


def test_return_format():
    """Test that RETURN section documents the required fields."""
    import otel_node_count

    assert 'nodes:' in otel_node_count.RETURN
    assert 'node_summary:' in otel_node_count.RETURN
    assert 'warnings:' in otel_node_count.RETURN
    assert 'total_count:' in otel_node_count.RETURN
    assert 'unique_host_ids:' in otel_node_count.RETURN
    assert 'duplicate_detections:' in otel_node_count.RETURN


def test_source_type_choices():
    """Test that source types are documented correctly."""
    import otel_node_count

    doc = otel_node_count.DOCUMENTATION
    assert 'opamp' in doc
    assert 'collector_api' in doc
    assert 'inventory' in doc
    assert 'kubernetes' in doc
