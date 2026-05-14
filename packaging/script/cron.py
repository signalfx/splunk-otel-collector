#!/usr/bin/env python3


import schedule
import time


from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import (
    OTLPMetricExporter,
)

from opentelemetry.metrics import (
    set_meter_provider,
)

from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader

import glob
import sys
import importlib
import os
from os.path import dirname, basename, join

jobs = os.environ.get('SCRIPT_JOBS_FOLDER', join(dirname(__file__), 'jobs'))

exporter = OTLPMetricExporter(insecure=True)
reader = PeriodicExportingMetricReader(exporter, export_interval_millis = int(os.environ.get("SCRIPT_EXPORT_INTERVAL", 5000)))
provider = MeterProvider(metric_readers=[reader])
set_meter_provider(provider)

modules = glob.glob(join(jobs, "*.py"))

sys.path.append(jobs)
for m in modules:
    importlib.import_module(basename(basename(m)[:-3]))

try:
    while True:
        schedule.run_pending()
        time.sleep(1)
except KeyboardInterrupt:
    schedule.clear()
