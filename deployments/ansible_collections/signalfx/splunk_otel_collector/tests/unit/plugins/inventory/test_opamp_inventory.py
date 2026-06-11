#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright 2025 Cisco Systems, Inc. and/or its affiliates
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

"""Tests for opamp inventory plugin.

These are basic tests that validate the plugin can be imported
and has the correct structure.
"""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import os
import sys

import pytest

# Add module paths
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/inventory'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../../plugins/module_utils'))


def test_inventory_plugin_imports():
    """Test that the inventory plugin can be imported."""
    from opamp import InventoryModule
    assert InventoryModule is not None


def test_inventory_plugin_name():
    """Test that the plugin has the correct NAME attribute."""
    from opamp import InventoryModule
    assert hasattr(InventoryModule, 'NAME')
    assert InventoryModule.NAME == 'signalfx.splunk_otel_collector.opamp'


def test_inventory_plugin_has_verify_file():
    """Test that plugin has verify_file method."""
    from opamp import InventoryModule
    plugin = InventoryModule()
    assert hasattr(plugin, 'verify_file')
    assert callable(plugin.verify_file)


def test_inventory_plugin_has_parse():
    """Test that plugin has parse method."""
    from opamp import InventoryModule
    plugin = InventoryModule()
    assert hasattr(plugin, 'parse')
    assert callable(plugin.parse)


def test_verify_file_valid():
    """Test verify_file accepts *.opamp.yml files when file exists."""
    import tempfile
    from opamp import InventoryModule
    plugin = InventoryModule()

    # Create temporary files to test with
    with tempfile.NamedTemporaryFile(suffix='.opamp.yml', delete=False) as f:
        temp_file = f.name
        f.write(b'plugin: cisco.splunk_otel_collector.opamp\n')

    try:
        # Test with real file
        assert plugin.verify_file(temp_file) is True
    finally:
        os.unlink(temp_file)


def test_verify_file_invalid():
    """Test verify_file rejects non-opamp files."""
    from opamp import InventoryModule
    plugin = InventoryModule()

    # Invalid files
    assert plugin.verify_file('inventory.yml') is False
    assert plugin.verify_file('hosts') is False
    assert plugin.verify_file('opamp.txt') is False


def test_inventory_plugin_documentation():
    """Test that plugin has proper documentation."""
    from opamp import InventoryModule, DOCUMENTATION
    assert DOCUMENTATION is not None
    assert 'name: opamp' in DOCUMENTATION
    assert 'plugin_type: inventory' in DOCUMENTATION


def test_inventory_plugin_examples():
    """Test that plugin has examples."""
    from opamp import InventoryModule, EXAMPLES
    assert EXAMPLES is not None
    assert 'plugin:' in EXAMPLES
    assert 'server_url:' in EXAMPLES
