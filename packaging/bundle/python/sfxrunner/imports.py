import importlib
import sys


def load_python_module(sys_paths, import_name):
    """
    Imports a Python module by name.  This will only have effect the first time
    it is called and subsequent calls will do nothing.
    """
    assert isinstance(sys_paths, (tuple, list)), "%s is not a list or tuple" % sys_paths

    for path in reversed(sys_paths):
        if path not in sys.path:
            sys.path.insert(1, path)

    return importlib.import_module(import_name)
