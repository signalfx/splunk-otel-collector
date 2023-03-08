import logging

from sfxmonitor import datapoint
from sfxrunner.scheduler.simple import SimpleScheduler

# Optional logger that you can use if you want to log messages back to the
# agent.  You can also use plain `print` statements, but they will go to the
# agent's ERROR log level, so logger is better for non-error messages.
logger = logging.getLogger(__name__)


class Monitor(object):
    """
    This is a full monitor implementation.  This gives you the most control
    over the operation of the monitor.  You can fire up multiple threads here
    or do whatever you want.  The main constraint is that you should *not* send
    datapoints from the same thread that the `configure` method is called on.

    If you don't need the flexibility that this type of custom Python monitor
    offers, you can use the simpler method of defining a monitor where you just
    define a `run(config, output)` function in a Python module, which will be
    called every `intervalSeconds` seconds that is specified in your monitor
    config.
    """

    def __init__(self, output):
        """
        The `output` instance is what you use to send datapoints back to the
        agent.
        """
        self.output = output
        self.scheduler = SimpleScheduler()

    def configure(self, config):
        """
        This method gets called immediately after __init__ with the
        configuration that was specified in the agent.  `config` is basically
        just a dictionary that corresponds to the montior config block in the
        agent, e.g.:

        {
          'MonitorID': '1',
          'configEndpointMappings': None,
          'discoveryRule': '',
          'disableHostDimensions': False,
          'pythonPath': None,
          'intervalSeconds': 10,
          'disableEndpointDimensions': False,
          'host': '10.5.2.0',  # Automatically populated if using auto-discovery
          'port': 8080,        # Automatically populated if using auto-discovery
          'scriptFilePath': '/usr/src/signalfx-agent/python/demoapp/monitor_complex.py',
          'config': None,
          'extraDimensions' : None,
          'type': 'python-monitor',
          'metricsToExclude': []
        }

        A lot of these config options are only used by the main agent process
        but are passed here for full transparency.  Any extra config options
        specified on the `python-monitor` monitor config will also be included
        in the top level of this dictionary.

        This method should return quickly (i.e. not block on I/O), and any real
        work should be done in separate threads.  As a convenience, there are
        two convenience scheduler classes available for you to use:
         - sfxrunner.scheduler.simple.SimpleScheduler (one thread per task)
         - sfxrunner.scheduler.interval.IntervalScheduler (shared threadpool)

        You are free to schedule tasks however you want, including using async
        event loops.  The main thing to consider is that you should not use the
        `output` object from the same thread that this method is called on, as
        that thread is used to manage the monitor instance.

        The `config` passed here will include an `intervalSeconds` option that
        every monitor's config has.  You should attempt to respect that if
        applicable.
        """
        logger.info(config)

        self.scheduler.run_on_interval(config.get("intervalSeconds"), self.gather)

    def gather(self):
        """
        This is just a private method on this particular monitor implementation
        that is used to actually do the monitoring work.
        """
        logger.info("Sending datapoint")
        self.output.send_datapoint(datapoint.gauge("my.gauge", 1, {"a": "1"}))

    def shutdown(self):
        """
        This method will be called, if defined, when the monitor should stop
        reporting.  It should stop any long-running background processing, such
        as the scheduler.
        """
        self.scheduler.stop()
