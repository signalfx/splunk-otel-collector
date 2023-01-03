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

void load_config(logger log, struct config *cfg, char *file_name) {
    read_config_file(log, cfg, file_name);
    if (cfg->service_name == NULL) {
        log_debug(log, "service_name not specified in config");
    }
    if (cfg->java_agent_jar == NULL) {
        log_debug(log, "java_agent_jar not specified in config");
    }
    if (cfg->resource_attributes == NULL) {
        log_debug(log, "resource_attributes not specified in config");
    }
    if (cfg->disable_telemetry == NULL) {
        log_debug(log, "disable_telemetry not specified in config");
    }
    if (cfg->generate_service_name == NULL) {
        log_debug(log, "generate_service_name not specified in config");
    }
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
        }
    }
}

void split_on_eq(char *string, struct kv *pair) {
    pair->k = strsep(&string, "=");
    pair->v = string;
}

bool str_eq_true(char *v) {
    return v != NULL && !streq("false", v) && !streq("FALSE", v) && !streq("0", v);
}

bool str_eq_false(char *v) {
    return v != NULL && (streq("false", v) || streq("FALSE", v) || streq("0", v));
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
}
