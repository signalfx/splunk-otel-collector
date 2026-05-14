from schedule import repeat, every
from opentelemetry.metrics import get_meter_provider

meter = get_meter_provider().get_meter("test")

# Gauge
gauge = meter.create_gauge("gauge")
val = 0
@repeat(every(3).seconds)
def increment():
    global val
    global gauge
    val += 3
    gauge.set(val)