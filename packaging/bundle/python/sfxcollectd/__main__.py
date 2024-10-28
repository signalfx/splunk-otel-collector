"""
Runs a collectd runner that can run a single collectd plugin instance.
"""


import logging
import logging.config

from sfxrunner.logs import PipeLogHandler, log_exc_traceback_as_error
from sfxrunner.messages import setup_io_pipes

from .runner import Runner

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)


def run():
    """
    Sets everything up and runs the adapter, blocking until it is shutdown by a
    shutdown message.
    """
    input_reader, output_writer = setup_io_pipes()

    # Logs go through our stdout pipe back to the agent.
    logger.addHandler(PipeLogHandler(output_writer))

    runner = Runner(input_reader, output_writer)
    logger.info("Starting up Collectd Python runner")
    runner.process()


try:
    run()
except Exception as e:  # pylint: disable=broad-except
    # runner.stop()
    log_exc_traceback_as_error()
