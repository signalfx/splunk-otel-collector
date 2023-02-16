#ifndef SPLUNK_INSTRUMENTATION_SPLUNK_H
#define SPLUNK_INSTRUMENTATION_SPLUNK_H

#include "logger.h"
#include "config.h"
#include "cmdline_reader.h"
#include "metrics_client.h"

static char *const disable_env_var = "DISABLE_SPLUNK_AUTOINSTRUMENTATION";
static char *const java_tool_options_var = "JAVA_TOOL_OPTIONS";
static char *const otel_service_name_var = "OTEL_SERVICE_NAME";
static char *const resource_attributes_var = "OTEL_RESOURCE_ATTRIBUTES";

typedef int (*has_access_func_t)(const char *);

typedef void (*load_config_func_t)(logger log, struct config *, char *);

void auto_instrument(
        logger log,
        has_access_func_t has_access,
        const char *program_name,
        load_config_func_t load_config_func,
        cmdline_reader cr,
        send_otlp_metric_func_t send_otlp_metric_func
);

int streq(const char *expected, const char *actual);

int concat_strings(char *dest, char *src, int tot_dest_size);

#endif //SPLUNK_INSTRUMENTATION_SPLUNK_H
