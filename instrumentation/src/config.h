#ifndef INSTRUMENTATION_CONFIG_H
#define INSTRUMENTATION_CONFIG_H

#include <stdbool.h>

#include "logger.h"

struct config {
    char *java_agent_jar;
    char *service_name;
    char *resource_attributes;
    char *disable_telemetry;
    char *generate_service_name;
    char *enable_profiler;
    char *enable_profiler_memory;
    char *enable_metrics;
};

void load_config(logger log, struct config *cfg, char *file_name);

bool str_to_bool(char *v, bool);

void free_config(struct config *cfg);

#endif //INSTRUMENTATION_CONFIG_H
