"""
Logic around the actual runner that manages the lifecycle of the monitor.
"""
import logging
import os
import threading

from sfxrunner.imports import load_python_module
from sfxrunner.logs import log_exc_traceback_as_error
from sfxrunner.messages import MSG_TYPE_CONFIGURE, MSG_TYPE_CONFIGURE_RESULT, MSG_TYPE_SHUTDOWN

from .output import Output
from .simple import SimpleMonitor

logger = logging.getLogger(__name__)


class Runner(object):  # pylint: disable=too-few-public-methods
    """
    Manages the creation and shutdown of the collectd plugin.  This acts as the
    mediator between the agent and the monitor proxy in Python.
    """

    def __init__(self, input_reader, output_writer):
        self.input_reader = input_reader
        self.output_writer = output_writer

    def process(self):
        """
        This is a fairly simple state machine.  It waits for a configure
        message that gives it the configuration and then it waits for a
        shutdown message, at which point this method returns.
        """
        logger.info("Waiting for configure message from agent")

        msg = self.input_reader.recv_msg()
        assert msg.type == MSG_TYPE_CONFIGURE, "Expected first message to be configure message"

        ready_event = threading.Event()
        output = Output(self.output_writer, ready_event)
        err = None
        try:
            config_func, shutdown_func = load_monitor(msg.payload, output)
            config_func(msg.payload)
        except Exception as e:  # pylint: disable=broad-except
            log_exc_traceback_as_error()
            err = e

        self.output_writer.send_msg(MSG_TYPE_CONFIGURE_RESULT, {"error": repr(err) if err else None})

        if err:
            return

        ready_event.set()
        msg = self.input_reader.recv_msg()
        assert msg.type == MSG_TYPE_SHUTDOWN, "Expected shutdown message, got %d" % msg.type

        if shutdown_func:
            shutdown_func()


def load_monitor(config, output):
    python_path = config.get("pythonPath") or []
    script_file_path = config.get("scriptFilePath")
    module_dir, module_file = os.path.split(script_file_path)
    if module_dir not in python_path:
        python_path.insert(0, module_dir)

    logger.debug("Appending %s to Python path", python_path)
    mod = load_python_module(python_path, os.path.splitext(module_file)[0])
    mon_cls = getattr(mod, "Monitor", None)
    if mon_cls:
        logger.info("Loaded class Monitor from %s", script_file_path)
        inst = mon_cls(output)
    else:
        logger.info("Loaded 'run' function from %s", script_file_path)
        run_func = getattr(mod, "run", None)
        if not run_func:
            raise ValueError(
                "Could not fine either a 'run' function or 'Montior' class in Python script '%s'" % (script_file_path,)
            )
        inst = SimpleMonitor(run_func, output)

    if not hasattr(inst, "configure"):
        raise ValueError("Monitor class in %s doesn't have a configure method" % script_file_path)
    return [getattr(inst, "configure"), getattr(inst, "shutdown", None)]
