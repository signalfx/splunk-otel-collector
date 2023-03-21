#!/usr/bin/env python
# coding=utf-8
#
# upstat-collectd

import sys

class UpStat():
    def __init__(self, interval=2, count=2):
        self.interval = interval
        self.count = count

    def get_upstats(self):
        return {"up": "1"}


class UpMon():
    def __init__(self):
        self.plugin_name = 'upstat'
        self.interval = 10.0
        self.include = set([])

    def configure_callback(self, conf):
        collectd.register_read(self.read_callback, self.interval)

    def dispatch_value(self, val_type, type_instance, value):
        """
        Dispatch a value to collectd
        """
        plugin_instance = "upstat"
        collectd.info('%s plugin: %s' % (self.plugin_name, 'Sending value: %s-%s.%s=%s' % (
            self.plugin_name,
            plugin_instance,
            '-'.join([val_type, type_instance]),
            value)))

        val = collectd.Values()
        val.plugin = self.plugin_name
        val.plugin_instance = plugin_instance
        val.type = val_type
        if len(type_instance):
            val.type_instance = type_instance
        val.values = [value]
        val.meta = {'0': True}
        val.dispatch()

    def read_callback(self):
        """
        Collectd read callback
        """
        collectd.info('Read callback called')
        upstat = UpStat(interval=self.interval)
        ds = upstat.get_upstats()

        if not ds:
            collectd.info('%s plugin: No info received.' % self.plugin_name)
            return

        for name in ds:
            self.dispatch_value(name, '', ds[name])



if __name__ == '__main__':
    upstat = UpStat()
    ds = upstat.get_upstats()

    for metric in ds:
        print("%s:%s" % (metric, ds[metric]))

    sys.exit(0)
else:
    import collectd

    upmon = UpMon()

    # Register callbacks
    collectd.register_config(upmon.configure_callback)