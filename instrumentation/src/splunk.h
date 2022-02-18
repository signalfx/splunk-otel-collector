#ifndef SPLUNK_INSTRUMENTATION_SPLUNK_H
#define SPLUNK_INSTRUMENTATION_SPLUNK_H

#include <stdbool.h>
#include "logger.h"
#include "config.h"

static char *const conf_file = "/usr/lib/splunk-instrumentation/auto-instr.conf";

static char *const disable_env_var_name = "DISABLE_SPLUNK_AUTOINSTRUMENTATION";

typedef bool (*has_access_func_t)(const char *);

typedef void (*load_config_func_t)(logger log, struct config *, char *);

void auto_instrument(logger log, has_access_func_t has_access, const char *program_name, load_config_func_t load_config_func);

bool streq(const char *expected, const char *actual);

#endif //SPLUNK_INSTRUMENTATION_SPLUNK_H
