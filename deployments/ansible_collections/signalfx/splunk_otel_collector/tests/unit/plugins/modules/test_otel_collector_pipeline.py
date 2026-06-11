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

"""Tests for otel_collector_pipeline module.

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
    import otel_collector_pipeline
    assert hasattr(otel_collector_pipeline, 'run_module')
    assert hasattr(otel_collector_pipeline, 'main')


def test_module_documentation():
    """Test that module has proper documentation."""
    import otel_collector_pipeline
    assert hasattr(otel_collector_pipeline, 'DOCUMENTATION')
    assert hasattr(otel_collector_pipeline, 'EXAMPLES')
    assert hasattr(otel_collector_pipeline, 'RETURN')

    # Verify key fields in documentation
    assert 'module: otel_collector_pipeline' in otel_collector_pipeline.DOCUMENTATION
    assert 'name:' in otel_collector_pipeline.DOCUMENTATION
    assert 'state:' in otel_collector_pipeline.DOCUMENTATION
    assert 'receivers:' in otel_collector_pipeline.DOCUMENTATION
    assert 'processors:' in otel_collector_pipeline.DOCUMENTATION
    assert 'exporters:' in otel_collector_pipeline.DOCUMENTATION


def test_module_examples():
    """Test that module has examples."""
    import otel_collector_pipeline
    assert 'otel_collector_pipeline:' in otel_collector_pipeline.EXAMPLES
    assert 'name:' in otel_collector_pipeline.EXAMPLES
    assert 'state:' in otel_collector_pipeline.EXAMPLES


def test_module_returns():
    """Test that module documents return values."""
    import otel_collector_pipeline
    assert 'changed:' in otel_collector_pipeline.RETURN
    assert 'pipeline:' in otel_collector_pipeline.RETURN
    assert 'diff:' in otel_collector_pipeline.RETURN
    assert 'message:' in otel_collector_pipeline.RETURN
