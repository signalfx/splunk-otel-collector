import logging
import time
import types
from functools import partial as p
from threading import Event, Thread

from sfxrunner.logs import log_exc_traceback_as_error

logger = logging.getLogger()


class SimpleScheduler(object):
    """
    This is a very simple scheduler that simply runs each item on a dedicated
    thread.  It is fine if the number of scheduled items is small and they will
    be run at about the same time.
    """

    def __init__(self):
        self.threads = []
        self.shutdown_event = Event()

    def run_on_interval(self, interval_in_seconds, func, immediately=True):
        """
        Calls func on a given interval.  Each func scheduled via this method
        may run in parallel with others if their intervals align.
        """
        thread = Thread(target=p(self._call_on_interval, interval_in_seconds, func, immediately))
        thread.daemon = True
        self.threads.append(thread)
        thread.start()

    def _call_on_interval(self, interval_in_seconds, func, immediately):
        next_run = time.time() + 0 if immediately else interval_in_seconds

        while True:
            # There is some inherent imprecision with this since there is no
            # guarantee that the sleep actually starts immediately after the
            # sleep duration is calculated, nor is the thread guaranteed to
            # wake up immediately after the duration, but this should be close
            # enough.  To get more precise, it could sleep for a time less than
            # needed and then do a busy spin until it reaches the desired time,
            # but even that is subject to some imprecision and would cause
            # significantly higher CPU usage.
            self.shutdown_event.wait(max(0, next_run - time.time()))
            if self.shutdown_event.is_set():
                return

            func_name = "<unknown>"
            if isinstance(func, types.FunctionType):
                func_name = f"{func.__module__}.{func.__name__}"
            elif isinstance(func, p):
                func_name = f"{func.func.__module__}.{func.func.__name__}"

            logger.debug("Running func %s", func_name)
            try:
                func()
            except Exception:  # pylint: disable=broad-except
                log_exc_traceback_as_error()
                # Swallow the exceptions after logging them.  We could
                # implement some kind of binary backoff logic like Collectd
                # uses.

            next_run += interval_in_seconds

    def stop(self):
        """
        Stops all existing threads and prevents any new ones from ever running.
        """
        self.shutdown_event.set()

        # Give the threads 5 seconds to shut down before returning
        wait_until = time.time() + 5
        for thr in self.threads:
            thr.join(max(0, wait_until - time.time()))
            if thr.is_alive():
                raise RuntimeError("Thread %s did not stop in time" % thr.ident)
