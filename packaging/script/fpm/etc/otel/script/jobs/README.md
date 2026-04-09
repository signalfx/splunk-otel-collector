This folder contains jobs running periodically.

You can create your own job by creating a Python file to run a script.

Counter example:
```python
from schedule import repeat, every
from opentelemetry.metrics import get_meter_provider

meter = get_meter_provider().get_meter("test")

# Counter
counter = meter.create_counter("counter")

@repeat(every(5).seconds)
def increment():
    counter.add(1)
```

Gauge example:
```python
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
```
