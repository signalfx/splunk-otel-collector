from functools import partial as p

from sfxrunner.scheduler.simple import SimpleScheduler


class SimpleMonitor(object):
    def __init__(self, run_func, output):
        self.run_func = run_func
        self.output = output
        self.scheduler = SimpleScheduler()

    def configure(self, config):
        self.scheduler.run_on_interval(config["intervalSeconds"], p(self.run_func, config, self.output))

    def shutdown(self):
        self.scheduler.stop()
