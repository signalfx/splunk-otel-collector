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

"""
Shared utility for reading and writing OpenTelemetry Collector YAML configuration files.
"""

from __future__ import absolute_import, division, print_function

__metaclass__ = type

import hashlib
import os
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

try:
    from ruamel.yaml import YAML
    HAS_RUAMEL_YAML = True
except ImportError:
    HAS_RUAMEL_YAML = False


class OtelConfig:
    """Manages OpenTelemetry Collector YAML configuration files."""

    def __init__(self, config_path: str):
        """Initialize OtelConfig with path to config file.

        Args:
            config_path: Path to the OTel collector config file
        """
        self.config_path = config_path
        self._data: Optional[Dict[str, Any]] = None
        self._original_hash: Optional[str] = None
        self._yaml = None

        if HAS_RUAMEL_YAML:
            self._yaml = YAML()
            self._yaml.preserve_quotes = True
            self._yaml.default_flow_style = False

    def load(self) -> Dict[str, Any]:
        """Load YAML config file. Return empty structure if file doesn't exist.

        Returns:
            Dictionary containing the config data
        """
        if not HAS_RUAMEL_YAML:
            raise ImportError("ruamel.yaml is required but not installed")

        config_file = Path(self.config_path)

        if not config_file.exists():
            # Return empty structure
            self._data = {
                'receivers': {},
                'processors': {},
                'exporters': {},
                'service': {
                    'pipelines': {}
                }
            }
            self._original_hash = self._compute_hash(self._data)
            return self._data

        with open(self.config_path, 'r', encoding='utf-8') as f:
            self._data = self._yaml.load(f)

        # Ensure required sections exist
        if self._data is None:
            self._data = {}

        for section in ['receivers', 'processors', 'exporters']:
            if section not in self._data:
                self._data[section] = {}

        if 'service' not in self._data:
            self._data['service'] = {}
        if 'pipelines' not in self._data['service']:
            self._data['service']['pipelines'] = {}

        self._original_hash = self._compute_hash(self._data)
        return self._data

    def save(self, backup: bool = False) -> None:
        """Write config back to file. Create timestamped backup if requested.

        Args:
            backup: If True, create a timestamped backup before writing
        """
        if not HAS_RUAMEL_YAML:
            raise ImportError("ruamel.yaml is required but not installed")

        if self._data is None:
            raise ValueError("No config data loaded")

        config_file = Path(self.config_path)

        # Create backup if requested and file exists
        if backup and config_file.exists():
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            backup_path = f"{self.config_path}.{timestamp}.bak"
            config_file.rename(backup_path)

        # Ensure directory exists
        config_file.parent.mkdir(parents=True, exist_ok=True)

        # Write the config
        with open(self.config_path, 'w', encoding='utf-8') as f:
            self._yaml.dump(self._data, f)

        # Update hash after save
        self._original_hash = self._compute_hash(self._data)

    def get_component(self, section: str, name: str) -> Optional[Dict[str, Any]]:
        """Get a component config from receivers/processors/exporters section.

        Args:
            section: One of 'receivers', 'processors', 'exporters'
            name: Name of the component

        Returns:
            Component configuration dict or None if not found
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if section not in self._data:
            return None

        return self._data[section].get(name)

    def set_component(self, section: str, name: str, config: Dict[str, Any]) -> None:
        """Set a component in receivers/processors/exporters section.

        Args:
            section: One of 'receivers', 'processors', 'exporters'
            name: Name of the component
            config: Configuration dictionary for the component
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if section not in ['receivers', 'processors', 'exporters']:
            raise ValueError(f"Invalid section: {section}")

        if section not in self._data:
            self._data[section] = {}

        self._data[section][name] = config

    def remove_component(self, section: str, name: str) -> bool:
        """Remove a component. Also remove from any pipeline references.

        Args:
            section: One of 'receivers', 'processors', 'exporters'
            name: Name of the component to remove

        Returns:
            True if component was removed, False if it didn't exist
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if section not in self._data or name not in self._data[section]:
            return False

        # Remove the component
        del self._data[section][name]

        # Remove from pipeline references
        pipeline_key = section  # receivers, processors, exporters
        if 'service' in self._data and 'pipelines' in self._data['service']:
            for pipeline_name, pipeline_config in self._data['service']['pipelines'].items():
                if pipeline_key in pipeline_config:
                    if name in pipeline_config[pipeline_key]:
                        pipeline_config[pipeline_key].remove(name)

        return True

    def get_pipeline(self, name: str) -> Optional[Dict[str, Any]]:
        """Get a pipeline from service.pipelines section.

        Args:
            name: Name of the pipeline

        Returns:
            Pipeline configuration dict or None if not found
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if 'service' not in self._data or 'pipelines' not in self._data['service']:
            return None

        return self._data['service']['pipelines'].get(name)

    def set_pipeline(
        self,
        name: str,
        receivers: List[str],
        processors: List[str],
        exporters: List[str]
    ) -> None:
        """Create or update a pipeline in service.pipelines.

        Args:
            name: Name of the pipeline
            receivers: List of receiver names
            processors: List of processor names
            exporters: List of exporter names
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if 'service' not in self._data:
            self._data['service'] = {}
        if 'pipelines' not in self._data['service']:
            self._data['service']['pipelines'] = {}

        self._data['service']['pipelines'][name] = {
            'receivers': receivers,
            'processors': processors,
            'exporters': exporters
        }

    def remove_pipeline(self, name: str) -> bool:
        """Remove a pipeline from service.pipelines.

        Args:
            name: Name of the pipeline to remove

        Returns:
            True if pipeline was removed, False if it didn't exist
        """
        if self._data is None:
            raise ValueError("Config not loaded. Call load() first.")

        if 'service' not in self._data or 'pipelines' not in self._data['service']:
            return False

        if name not in self._data['service']['pipelines']:
            return False

        del self._data['service']['pipelines'][name]
        return True

    def has_changed(self) -> bool:
        """Check if config differs from what was loaded.

        Returns:
            True if config has been modified
        """
        if self._data is None:
            return False

        current_hash = self._compute_hash(self._data)
        return current_hash != self._original_hash

    def _compute_hash(self, data: Dict[str, Any]) -> str:
        """Compute hash of config data for change detection.

        Args:
            data: Configuration dictionary

        Returns:
            SHA256 hash of the data as a hex string
        """
        import json
        # Use json.dumps with sorted keys for consistent hashing
        data_str = json.dumps(data, sort_keys=True)
        return hashlib.sha256(data_str.encode('utf-8')).hexdigest()
