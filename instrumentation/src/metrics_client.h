#ifndef INSTRUMENTATION_METRICS_CLIENT_H
#define INSTRUMENTATION_METRICS_CLIENT_H

#include "logger.h"

typedef void (*send_otlp_metric_func_t)(logger, char *);

void send_otlp_metric(logger, char *);

#endif //INSTRUMENTATION_METRICS_CLIENT_H
