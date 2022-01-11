#ifndef SPLUNK_INSTRUMENTATION_SPLUNK_H
#define SPLUNK_INSTRUMENTATION_SPLUNK_H

#include <stdbool.h>
#include "logger.h"

static char *const DISABLE_ENV_VAR_NAME = "DISABLE_SPLUNK_AUTOINSTRUMENTATION";

typedef bool (*has_access_t)(const char *);

void auto_instrument(logger log, has_access_t has_access, const char *program_name, const char *jar_path);

bool has_read_access(const char *s);

bool is_disable_env_set();

#endif //SPLUNK_INSTRUMENTATION_SPLUNK_H
