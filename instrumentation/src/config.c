#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "config.h"
#include "splunk.h"

struct kv {
    char *k;
    char *v;
};

void read_config_file(logger log, struct config *cfg, char *file_name);

void read_lines(struct config *cfg, FILE *fp);

void split_on_eq(char *string, struct kv *pair);

void log_config_field(logger log, const char *field, const char *value);

void load_config(logger log, struct config *cfg, char *file_name) {
    read_config_file(log, cfg, file_name);
    log_config_field(log, "service_name", cfg->service_name);
    log_config_field(log, "java_agent_jar", cfg->java_agent_jar);
    log_config_field(log, "resource_attributes", cfg->resource_attributes);
    log_config_field(log, "disable_telemetry", cfg->disable_telemetry);
    log_config_field(log, "generate_service_name", cfg->generate_service_name);
    log_config_field(log, "enable_profiler", cfg->enable_profiler);
    log_config_field(log, "enable_profiler_memory", cfg->enable_profiler_memory);
    log_config_field(log, "enable_metrics", cfg->enable_metrics);
}

void log_config_field(logger log, const char *field, const char *value) {
    char msg[MAX_LOG_LINE_LEN] = "";
    if (value == NULL) {
        snprintf(msg, MAX_LOG_LINE_LEN, "config: %s not specified", field);
    } else {
        snprintf(msg, MAX_LOG_LINE_LEN, "config: %s=%s", field, value);
    }
    log_debug(log, msg);
}

void read_config_file(logger log, struct config *cfg, char *file_name) {
    FILE *fp = fopen(file_name, "r");
    char log_line[MAX_LOG_LINE_LEN];
    if (fp == NULL) {
        strcpy(log_line, "file not found: ");
        strncat(log_line, file_name, MAX_LOG_LINE_LEN - strlen(log_line) - 1);
        log_debug(log, log_line);
        return;
    }

    strcpy(log_line, "reading config file: ");
    strncat(log_line, file_name, MAX_LOG_LINE_LEN - strlen(log_line) - 1);
    log_debug(log, log_line);
    read_lines(cfg, fp);
    fclose(fp);
}

void read_lines(struct config *cfg, FILE *fp) {
    static const int buflen = 255;
    char buf[buflen];
    struct kv pair = {.k = NULL, .v = NULL};
    while ((fgets(buf, buflen, fp)) != NULL) {
        buf[strcspn(buf, "\n")] = 0;
        split_on_eq(buf, &pair);
        if (streq(pair.k, "java_agent_jar")) {
            cfg->java_agent_jar = strdup(pair.v);
        } else if (streq(pair.k, "service_name")) {
            cfg->service_name = strdup(pair.v);
        } else if (streq(pair.k, "resource_attributes")) {
            cfg->resource_attributes = strdup(pair.v);
        } else if (streq(pair.k, "disable_telemetry")) {
            cfg->disable_telemetry = strdup(pair.v);
        } else if (streq(pair.k, "generate_service_name")) {
            cfg->generate_service_name = strdup(pair.v);
        } else if (streq(pair.k, "enable_profiler")) {
            cfg->enable_profiler = strdup(pair.v);
        } else if (streq(pair.k, "enable_profiler_memory")) {
            cfg->enable_profiler_memory = strdup(pair.v);
        } else if (streq(pair.k, "enable_metrics")) {
            cfg->enable_metrics = strdup(pair.v);
        }
    }
}

void split_on_eq(char *string, struct kv *pair) {
    pair->k = strsep(&string, "=");
    pair->v = string;
}

int str_to_bool(char *v, int defaultVal) {
    if (v == NULL) {
        return defaultVal;
    }
    if (streq("false", v) || streq("FALSE", v) || streq("0", v)) {
        return 0;
    }
    return 1;
}

void free_config(struct config *cfg) {
    if (cfg->java_agent_jar != NULL) {
        free(cfg->java_agent_jar);
    }
    if (cfg->service_name != NULL) {
        free(cfg->service_name);
    }
    if (cfg->resource_attributes != NULL) {
        free(cfg->resource_attributes);
    }
    if (cfg->disable_telemetry != NULL) {
        free(cfg->disable_telemetry);
    }
    if (cfg->generate_service_name != NULL) {
        free(cfg->generate_service_name);
    }
    if (cfg->enable_profiler != NULL) {
        free(cfg->enable_profiler);
    }
    if (cfg->enable_profiler_memory != NULL) {
        free(cfg->enable_profiler_memory);
    }
    if (cfg->enable_metrics != NULL) {
        free(cfg->enable_metrics);
    }
}
