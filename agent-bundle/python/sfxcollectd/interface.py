"""
Our implementation of the collectd Python interface
"""
import logging
import sys
import time
import types
from functools import partial

logger = logging.getLogger()


class CollectdInterface(object):
    """
    This ultimately exposes and implements all of the collectd interface to the
    Python plugins.

    It can only handle a single plugin and unified set of plugin configuration.
    """

    def __init__(self, scheduler, default_interval):
        self.default_interval = default_interval
        self.config_callback = None
        self.read_initializers = []
        self.shutdown_callbacks = []
        self.scheduler = scheduler
        self.names = set()

    def register_config(self, callback):
        """
        Implementation of the register_config function from collectd
        """
        if self.config_callback:
            logger.warning("Config callback was already registered, re-registering")

        self.config_callback = callback

    def register_read(self, callback, interval=None, data=None, name=None):
        """
        Implementation of the register_read function, immediately schedules the
        read callback to run on the determined interval.
        """
        final_name = name or "%s.%s" % (callback.__module__, callback.__name__)
        if final_name in self.names:
            logger.error("Read callback name %s already registered, registering twice", final_name)
        self.names.add(final_name)

        if data:
            func = partial(callback, data)
        else:
            func = callback

        self.read_initializers.append(
            partial(self.scheduler.run_on_interval, interval or self.default_interval, func, immediately=True)
        )

    @staticmethod
    def register_init(callback):
        """
        Just go ahead and call the init callback right away when registered
        since we don't have any interpreter setup to worry about by that point
        """
        callback()

    def register_shutdown(self, callback):
        """
        Mimics the collectd-python register_shutdown method
        """
        self.shutdown_callbacks.append(callback)


class Values(object):  # pylint: disable=too-many-instance-attributes
    """
    Implementation of the Values object in collectd-python

    See https://collectd.org/documentation/manpages/collectd-python.5.shtml#values
    """

    # Slots improves memory efficiency of value lists
    __slots__ = [
        "type",
        "values",
        "host",
        "plugin",
        "plugin_instance",
        "time",
        "type_instance",
        "interval",
        "meta",
        "dsnames",
        "dstypes",
        "message",
        "severity",
    ]

    def __init__(  # pylint: disable=too-many-arguments
        self,
        type=None,  # pylint: disable=redefined-builtin
        values=None,
        host="",
        plugin=None,
        plugin_instance="",
        time=None,  # pylint: disable=redefined-outer-name
        type_instance="",
        interval=None,
        meta=None,
        # This gives us the ability to have python plugins that specify
        # their own types in code instead of being dependent on types.db
        dsnames=None,
        dstypes=None,
        message=None,
        severity=None,
    ):
        self.type = type
        self.values = values or []
        self.host = host
        self.plugin = plugin
        self.plugin_instance = plugin_instance
        self.time = time
        self.type_instance = type_instance
        self.interval = interval
        self.meta = meta or {}
        self.dsnames = dsnames
        self.dstypes = dstypes
        self.message = message
        self.severity = severity

    def dispatch(self):
        """
        This method technically can accept any of the kwargs accepted by the
        constructor but we'll defer supporting that until needed
        """
        if not self.time:
            self.time = time.time()
        logger.debug("Dispatching value %s", repr(self))
        # convert boolean values to their integer type
        # because that is what collectd does
        self.values = [int(value) if isinstance(value, bool) else value for value in self.values]
        Values._dispatcher_func(self)

    @classmethod
    def set_dispatcher_func(cls, func):
        """
        Called to inject the function that sends datapoints when the `dispatch`
        method is called on a Values instance
        """
        cls._dispatcher_func = func

    def __repr__(self):
        return (
            "{plugin: %s; plugin_instance: %s; type: %s; type_instance: %s; "
            "host: %s; values: %s; time: %s; meta: %s; dsnames: %s; "
            "dstypes: %s; message: %s; severity: %s}"
            % (
                self.plugin,
                self.plugin_instance,
                self.type,
                self.type_instance,
                self.host,
                self.values,
                self.time,
                self.meta,
                self.dsnames,
                self.dstypes,
                self.message,
                self.severity,
            )
        )


def inject_collectd_module(interface, send_values_func):
    """
    Creates and registers the collectd python module so that plugins can import
    it properly.  This should only be called once per python interpreter.
    """
    assert "collectd" not in sys.modules, "collectd module should only be created once"

    mod = types.ModuleType("collectd")
    mod.register_config = interface.register_config
    mod.register_init = interface.register_init
    mod.register_read = interface.register_read
    mod.register_shutdown = interface.register_shutdown
    mod.info = logger.info
    mod.error = logger.error
    mod.warning = logger.warning
    mod.notice = logger.warning
    mod.info = logger.info
    mod.debug = logger.debug

    Values.set_dispatcher_func(send_values_func)
    mod.Values = Values
    sys.modules["collectd"] = mod
