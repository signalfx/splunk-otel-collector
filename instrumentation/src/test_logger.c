#include <stdlib.h>
#include <string.h>
#include "logger.h"

struct logger_impl {
    int i;
    char *logs[MAX_LOG_LINE_LEN];
};

logger new_logger() {
    logger l = malloc(sizeof *l);
    if (l) {
        l->i = 0;
    }
    return l;
}

void save_log(logger l, char *s) {
    l->logs[l->i++] = strdup(s);
}

void log_info(logger l, char *s) {
    save_log(l, s);
}

void log_debug(logger l, char *s) {
    save_log(l, s);
}

void log_warning(logger l, char *s) {
    save_log(l, s);
}

void free_logger(logger l) {
    for (int i = 0; i < l->i; ++i) {
        free(l->logs[i]);
    }
    free(l);
}

int get_logs(logger l, char *buf[]) {
    for (int i = 0; i < l->i; ++i) {
        buf[i] = l->logs[i];
    }
    return l->i;
}
