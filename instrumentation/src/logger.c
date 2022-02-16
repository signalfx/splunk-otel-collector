#include <stdlib.h>
#include <syslog.h>
#include "logger.h"

static char *const prefix = "splunk-instrumentation: %s";

struct logger_impl {
};

logger new_logger() {
    logger out = malloc(sizeof *out);
    return out;
}

void log_debug(logger l, char *s) {
    syslog(LOG_DEBUG, prefix, s);
}

void log_info(logger l, char *s) {
    syslog(LOG_INFO, prefix, s);
}

void log_warning(logger l, char *s) {
    syslog(LOG_WARNING, prefix, s);
}

void free_logger(logger l) {
    free(l);
}
