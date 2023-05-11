import threading
from itertools import groupby
from operator import attrgetter

from .datapoint import cumulative, gauge

# Message type for a datapoint list.
MSG_TYPE_DATAPOINT_LIST = 200


class Output(object):
    _datapoint_key_func = attrgetter("type")

    def __init__(self, output_writer, ready_event):
        self.output_writer = output_writer
        self.ready_event = ready_event
        self.creator_tid = threading.current_thread().ident

    def send_gauge(self, name, value, dimensions=None, timestamp=None):
        """
        Sends a gauge metric with the given `name` and `value`, and optional
        `dimensions`, which should be a dict of string keys and values.  If
        `timestamp` is left unspecified, it will default to the curent time.
        """
        return self.send_datapoint(gauge(name, value, dimensions, timestamp))

    def send_cumulative(self, name, value, dimensions=None, timestamp=None):
        """
        Sends a cumulative counter metric with the given `name` and `value`,
        and optional `dimensions`, which should be a dict of string keys and
        values.  If `timestamp` is left unspecified, it will default to the
        curent time.
        """
        return self.send_datapoint(cumulative(name, value, dimensions, timestamp))

    def send_datapoint(self, datapoint):
        """
        Convenince function that calls `send_datapoints` with the given
        `datapoint` wrapped in a list.
        """
        self.send_datapoints([datapoint])

    def send_datapoints(self, datapoints):
        """
        Sends a set of datapoints back up to the agent.  `datapoints` should be
        a list of `sfxmonitor.datapoint.Datapoint` instances.
        """
        # This prevents deadlock with the ready_event flag.
        if threading.current_thread().ident == self.creator_tid:
            raise RuntimeError(
                "You should not send datapoints from the same thread that the 'configure' method "
                "on your monitor was called"
            )

        # Don't send datapoints until the configure result has been sent back
        # to the agent.
        self.ready_event.wait()

        out = {}
        for typ, group in groupby(sorted(datapoints, key=self._datapoint_key_func), self._datapoint_key_func):
            out[typ] = [dp.as_dict() for dp in group]

        self.output_writer.send_msg(MSG_TYPE_DATAPOINT_LIST, out)
