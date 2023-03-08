"""
Logic around the actual runner that manages the lifecycle of the plugin.
"""
import logging

from sfxrunner.logs import log_exc_traceback_as_error
from sfxrunner.messages import MSG_TYPE_CONFIGURE, MSG_TYPE_CONFIGURE_RESULT, MSG_TYPE_SHUTDOWN

from .collectd import CollectdMonitorProxy

logger = logging.getLogger(__name__)

# Message type for a value list from the collectd plugin.  This encapsulates
# both metrics and events (notifications).
MSG_TYPE_VALUE_LIST = 100


class Runner(object):
    """
    Manages the creation and shutdown of the collectd plugin.  This acts as the
    mediator between the agent and the monitor proxy in Python.
    """

    def __init__(self, input_reader, output_writer):
        self._monitor_proxy = None

        self.input_reader = input_reader
        self.output_writer = output_writer

    def process(self):
        """
        This is a fairly simple state machine.  It waits for a configure
        message that gives it the configuration and then it waits for a
        shutdown message, at which point this method returns.
        """
        logger.info("Waiting for configure message")

        msg = self.input_reader.recv_msg()
        assert msg.type == MSG_TYPE_CONFIGURE, "Expected first message to be configure message"

        self._monitor_proxy = CollectdMonitorProxy(self.send_value_list)
        err = None
        try:
            # This should fire off a separate thread that runs the various
            # callbacks for the collectd plugin code.
            self._monitor_proxy.configure(msg.payload)
        except Exception as e:  # pylint: disable=broad-except
            log_exc_traceback_as_error()
            err = e

        self.output_writer.send_msg(MSG_TYPE_CONFIGURE_RESULT, {"error": repr(err) if err else None})

        if err:
            return

        self._monitor_proxy.start_reading()

        msg = self.input_reader.recv_msg()
        assert msg.type == MSG_TYPE_SHUTDOWN, "Expected shutdown message, got %d" % msg.type

        self._monitor_proxy.shutdown()

    def send_value_list(self, value_list):
        """
        Sends a value list from collectd to the agent by json encoding it
        """
        # output_writer is thread-safe, which is necessary since plugins can
        # register multiple read callbacks which could emit value lists
        # simultaneously.
        self.output_writer.send_msg(MSG_TYPE_VALUE_LIST, {x: getattr(value_list, x) for x in value_list.__slots__})
