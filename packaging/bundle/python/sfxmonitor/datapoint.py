from collections import namedtuple

TYPE_GAUGE = "gauge"
TYPE_COUNTER = "counter"
TYPE_CUMULATIVE = "cumulative_counter"


class Datapoint(namedtuple("Datapoint", ["name", "type", "value", "dimensions", "timestamp"])):
    def as_dict(self):
        return {
            "metric": self.name,
            "value": self.value,
            "dimensions": self.dimensions,
            "timestamp": int(self.timestamp * 1000) if self.timestamp else None,
        }


def gauge(name, value, dimensions=None, timestamp=None):
    """
    Creates a gauge datapoint.
    """
    return Datapoint(name, TYPE_GAUGE, value, dimensions, timestamp)


def cumulative(name, value, dimensions=None, timestamp=None):
    """
    Creates a cumulative counter datapoint.
    """
    return Datapoint(name, TYPE_CUMULATIVE, value, dimensions, timestamp)
