"""
Collectd's Python plugins are fairly difficult to work with.  There is no
first-class concept of multiple "instances" of a plugin, only one plugin module
that registers a single config callback that then gets invoked for each
configuration of the plugin.  Some plugins rely on a single read callback that
gets its configuration from a global variable and some create separate read
callbacks upon each invocation of the config callback.  Moreover, some plugins
don't handle multiple configurations at all and just overwrite the previous
configuration if the config callback is called multiple times.

There is also no concept of a plugin being "unconfigured", where a certain
configuration is removed.  There is a "shutdown" callback, but this is
dependent on the plugin implementing it properly. For these reasons, it is
virtually impossible to generically support dynamic monitor creation and
shutdown using the same Python interpreter.  Therefore, it is best if we simply
run each instance of the collectd-based monitor in a separate Python
interpreter, which means launching multiple instances of this Python adapter
process.

PEP 554 (https://www.python.org/dev/peps/pep-0554) could be useful for us if it
ever gets made a standard and we move to Python 3 since it would allow for
loading modules multiple times with totally isolated configurations of the same
plugin.  If a monitor associated with a particular config shutdown, then the
subinterpreter could just be destroyed.
"""


import logging
from collections import namedtuple

from sfxrunner.imports import load_python_module
from sfxrunner.scheduler.simple import SimpleScheduler

from .config import Config
from .interface import CollectdInterface, inject_collectd_module
from .typesdb import parse_types_db

logger = logging.getLogger(__name__)

DataSetCache = namedtuple("DataSetCache", "sources names types")


class CollectdMonitorProxy(object):
    """
    This is roughly analogous to a Monitor struct in the agent
    """

    def __init__(self, send_values_func):
        self.send_values_func = send_values_func

        # Our implementation of the collectd Python interface
        self.interface = None
        # The config that comes from the agent
        self.config = None
        self.scheduler = SimpleScheduler()
        self.datasets = {}

    def configure(self, monitor_config):
        """
        Calls the config callback with the given config, properly converted to
        the object that the collectd plugin expects.
        """
        assert (
            "pluginConfig" in monitor_config
        ), "Monitor config for collectd python should have a field called 'pluginConfig'"

        self.config = monitor_config
        self.interface = CollectdInterface(self.scheduler, monitor_config["intervalSeconds"])

        self.init_types_db_data_sets(monitor_config.get("typesDBPaths", []))

        inject_collectd_module(self.interface, self.send_value_list_with_dataset)

        module_paths = monitor_config.get("modulePaths", [])
        import_name = monitor_config["moduleName"]

        load_python_module(module_paths, import_name)

        if not self.interface.config_callback:
            raise RuntimeError("No config callback was registered, cannot configure")

        collectd_config = Config.from_monitor_config(monitor_config["pluginConfig"])

        self.interface.config_callback(collectd_config)

        if not self.interface.read_initializers:
            raise RuntimeError("No read callbacks were registered after configuring, this plugin is useless!")

    def start_reading(self):
        """
        Kicks off the read callback cycle
        """
        for read_init in self.interface.read_initializers:
            read_init()

    def init_types_db_data_sets(self, paths):
        """
        Parse out the typesdb files and populate a set that can be used to
        quickly look them up to attach them to value lists
        """
        logger.info("Loading types.db files: %s", paths)

        for types_db_path in paths:
            with open(types_db_path, "r", encoding="utf-8") as fd:
                for dataset in parse_types_db(fd.read()):
                    self.datasets[dataset.name] = DataSetCache(
                        sources=dataset.sources,
                        names=[s.name for s in dataset.sources],
                        types=[s.type for s in dataset.sources],
                    )

        logger.debug("Registered data sets: %s", self.datasets)

    def send_value_list_with_dataset(self, value_list):
        """
        Sets the dsnames and dstypes fields on the value list from the cache of
        data sets configured on the monitor
        """
        ds_cache = self.datasets.get(value_list.type)
        if not ds_cache:
            logger.error(
                "Type %s was not found in the types.db files configured (%s)",
                value_list.type,
                self.config.get("typesDBPaths"),
            )
            return

        value_list.dsnames = ds_cache.names
        value_list.dstypes = ds_cache.types

        self.send_values_func(value_list)

    def shutdown(self):
        """
        Called once when the monitor is supposed to shut down
        """
        if self.interface:
            self.scheduler.stop()

            for callback in self.interface.shutdown_callbacks:
                callback()
