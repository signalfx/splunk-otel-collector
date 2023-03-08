"""
Logic for scheduling reads across a shared threadpool
"""


import heapq
import logging
import sys
import time
from threading import Event, Lock, Thread


class IntervalScheduler(object):  # pylint: disable=too-many-instance-attributes
    """
    Facilitates executing a set of functions at some regular interval
    across a shared set of threads.

    This implementation handles adding and removing scheduled functions at any
    interval.
    """

    def __init__(self, max_thread_count=5):
        self.max_thread_count = max_thread_count
        self.threads = []

        self.heap = []
        self.heap_lock = Lock()

        self.stop_event = Event()

        # This is used to facilitate cancelling gatherings.
        self.func_blacklist = []

        # Event that can be triggered when a new item is scheduled for the
        # first time that needs to run before the next scheduled item in the
        # heap.
        self.new_earlier_event = Event()
        self.next_scheduled = sys.maxsize  # pylint: disable=no-member

    def _add_thread(self):
        if len(self.threads) >= self.max_thread_count:
            return
        thr = Thread(target=self._gather_metrics_thread)
        thr.daemon = True
        self.threads.append(thr)
        thr.start()

    def stop(self):
        """
        Stops the scheduler.
        """
        self.stop_event.set()
        # This is kind of a hack, but it awakens all threads so they stop
        # immediately
        self.new_earlier_event.set()

    def run_on_interval(self, interval_in_seconds, func, immediately=True):
        """
        @param immediately: whether to run the func immediately when registered
        or wait until `interval_in_seconds` for the first run
        """
        when = time.time() + (0 if immediately else interval_in_seconds)

        with self.heap_lock:
            is_earliest = self._schedule_gathering(when, func, interval_in_seconds)

            # This tests for an edge case where a new interval is supposed to
            # begin before any scheduled gatherings.  We have to awaken at
            # least one gather thread and have it reset to wait for the earlier
            # one.
            if is_earliest:
                self.new_earlier_event.set()

            if len(self.heap) < self.max_thread_count:
                self._add_thread()

        def cancel():
            if cancel.was_called:
                return

            with self.heap_lock:
                # The func should only be in either the heap once, or in a
                # single gather thread awaiting execution.
                for i, (_, heap_func, _) in enumerate(self.heap):
                    if heap_func == func:
                        del self.heap[i]
                        heapq.heapify(self.heap)
                        cancel.was_called = True
                        return

                # If the func wasn't in the heap, then it must be currently
                # scheduled gather thread awaiting execution.  This will
                # tell the gather thread to not run it again, nor
                # reschedule it, which will effectively stop the gathering.
                self.func_blacklist.append(func)
            logging.error("Could not find gather event on heap to cancel!")

        cancel.was_called = False

        return cancel

    def _schedule_gathering(self, when, func, interval_in_seconds):
        """
        Assumes caller holds heap lock!

        @returns: bool specifying whether the scheduled gather is supposed to
        occur earlier than the next scheduled gathering
        """
        heapq.heappush(self.heap, (when, func, interval_in_seconds))
        logging.debug("Inserted %s into heap: %s", (when, func, interval_in_seconds), self.heap)
        if when < self.next_scheduled:
            self.next_scheduled = when
            return True
        return False

    def _gather_metrics_thread(self):
        """
        This is the main function of the separate worker threads.
        """
        while True:
            if self.stop_event.is_set():
                return

            with self.heap_lock:
                try:
                    when, func, interval = heapq.heappop(self.heap)
                    self.next_scheduled = when
                except IndexError:
                    # There is nothing to do so shutdown this thread.  Another
                    # will be started up if we are under the max thread count
                    # and there are more scheduled than the current number of
                    # threads
                    logging.info("Nothing for gather thread to do, shutting down")
                    return

            # If _wait_until_gather returns false then it isn't time to
            # actually do the gathering because either 1) the scheduler is
            # shutting down, or 2) there is another item scheduler earlier than
            # what was previously scheduled -- so reschedule the current
            # gathering and start over.
            if not self._wait_until_gather(when):
                with self.heap_lock:
                    self._schedule_gathering(when, func, interval)
                    continue

            with self.heap_lock:
                if func in self.func_blacklist:
                    self.func_blacklist.remove(func)
                    continue

            func()

            with self.heap_lock:
                self._schedule_gathering(interval + when, func, interval)

    def _wait_until_gather(self, when):
        """
        Pauses the gather thread until either the gathering is supposed to happen or
        until a new earlier event was triggered, in which case all of the
        threads will wake up and one of them will reinsert their currently
        scheduled gathering onto the heap and take an earlier one.

        @return: whether `when` has been reached or not and the gather should
        occur. False indicates that there was a new earlier gathering that
        should happen and thus this thread should give up it's currently
        scheduled gathering.
        """
        secs_until_gather = when - time.time()

        while secs_until_gather > 0:
            self.new_earlier_event.wait(secs_until_gather)

            if self.stop_event.is_set():
                return True

            with self.heap_lock:
                # This means there was a more recent gather scheduled than what
                # we are currently waiting for.  In this case, we want to
                # push the scheduled gather back onto the heap and pull
                # another (which will be an earlier one).  Only one thread
                # will actually do this because of the heap lock.  The rest
                # will just go back to sleep with an updated wait time.
                if self.new_earlier_event.is_set():
                    self.new_earlier_event.clear()
                    return False

            secs_until_gather = when - time.time()
        return True
