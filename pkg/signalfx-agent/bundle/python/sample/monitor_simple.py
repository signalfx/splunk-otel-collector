import logging
import multiprocessing

# Optional logger that you can use if you want to log messages back to the
# agent.  You can also use plain `print` statements, but they will go to the
# agent's ERROR log level, so logger is better for non-error messages.
logger = logging.getLogger(__name__)


# `config` is a dict of the configuration provided to the agent, including any
# custom config options that are useful in this function.
# `output` is how you get datapoints back to the main agent process.
def run(config, output):
    # This will emit a gauge metric called "num_cores" with a value of the cpu
    # count of the current system, and with a single dimension called "source"
    # with the value "python".
    output.send_gauge("num_cores", multiprocessing.cpu_count(), {"source": "python"})

    # This shows getting a config option.  You can use `dict.get` to provide
    # default values and avoid crashes if the value isn't provided.
    if config.get("sendCounter", False):
        run.num_calls += 1
        # Sends a cumulative counter metric called "num_calls"
        output.send_cumulative("num_calls", run.num_calls)


# The simplest way to keep state between invocations of the "run" function is
# just to stick an attribute on it.  Only one instance of the monitor will live
# in this process so it won't be shared with anything else.
run.num_calls = 0
