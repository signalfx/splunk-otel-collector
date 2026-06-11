"""Root conftest for adding project root to Python path."""

import sys
from pathlib import Path

# Add project root to Python path so tests can import extensions module
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))
