#ifndef INSTRUMENTATION_METRICS_CLIENT_H
#define INSTRUMENTATION_METRICS_CLIENT_H

#include <stdbool.h>

#include "logger.h"

typedef void (*send_otlp_metric_func_t)(logger);

void send_otlp_metric(logger);

#endif //INSTRUMENTATION_METRICS_CLIENT_H
