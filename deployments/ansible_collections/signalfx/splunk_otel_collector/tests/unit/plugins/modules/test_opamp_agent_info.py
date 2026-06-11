#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

"""Tests for opamp_agent_info module.

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
    import opamp_agent_info
    assert hasattr(opamp_agent_info, 'run_module')
    assert hasattr(opamp_agent_info, 'main')


def test_module_documentation():
    """Test that module has proper documentation."""
    import opamp_agent_info
    assert hasattr(opamp_agent_info, 'DOCUMENTATION')
    assert hasattr(opamp_agent_info, 'EXAMPLES')
    assert hasattr(opamp_agent_info, 'RETURN')


def test_documentation_format():
    """Test that module documentation follows expected format."""
    import opamp_agent_info

    # Check that DOCUMENTATION contains key fields
    assert 'module: opamp_agent_info' in opamp_agent_info.DOCUMENTATION
    assert 'short_description:' in opamp_agent_info.DOCUMENTATION
    assert 'version_added:' in opamp_agent_info.DOCUMENTATION
    assert 'options:' in opamp_agent_info.DOCUMENTATION

    # Check that required options are documented
    assert 'server_url:' in opamp_agent_info.DOCUMENTATION
    assert 'token:' in opamp_agent_info.DOCUMENTATION


def test_examples_format():
    """Test that EXAMPLES section contains valid YAML-like content."""
    import opamp_agent_info

    assert 'cisco.splunk_otel_collector.opamp_agent_info' in opamp_agent_info.EXAMPLES or \
           'signalfx.splunk_otel_collector.opamp_agent_info' in opamp_agent_info.EXAMPLES
    assert 'server_url:' in opamp_agent_info.EXAMPLES
    assert 'token:' in opamp_agent_info.EXAMPLES


def test_return_format():
    """Test that RETURN section documents the agents list."""
    import opamp_agent_info

    assert 'agents:' in opamp_agent_info.RETURN
    assert 'description:' in opamp_agent_info.RETURN
    assert 'type:' in opamp_agent_info.RETURN
