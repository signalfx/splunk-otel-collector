from schedule import repeat, every
from opentelemetry.metrics import get_meter_provider

meter = get_meter_provider().get_meter("test")

# Counter
counter = meter.create_counter("counter")

@repeat(every(5).seconds)
def increment():
    counter.add(1)