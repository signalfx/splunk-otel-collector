#ifndef SPLUNK_INSTRUMENTATION_LOGGER_H
#define SPLUNK_INSTRUMENTATION_LOGGER_H

#define TEST_LOGS_MAX_LEN 256

typedef struct logger_impl *logger;

logger new_logger();

void log_debug(logger l, char *s);

void log_info(logger l, char *s);

int get_logs(logger l, char *buf[256]);

void free_logger(logger l);

#endif //SPLUNK_INSTRUMENTATION_LOGGER_H
