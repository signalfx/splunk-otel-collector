#include <stdlib.h>
#include "logger.h"

struct logger_impl {
    int i;
    char *logs[TEST_LOGS_MAX_LEN];
};

logger new_logger() {
    logger ref = malloc(sizeof *ref);
    if (ref) {
        ref->i = 0;
    }
    return ref;
}

void log_info(logger l, char *s) {
    l->logs[l->i++] = s;
}

void log_debug(logger l, char *s) {
    l->logs[l->i++] = s;
}

void free_logger(logger l) {
    free(l);
}

int get_logs(logger l, char *buf[]) {
    for (int i = 0; i < l->i; ++i) {
        buf[i] = l->logs[i];
    }
    return l->i;
}
