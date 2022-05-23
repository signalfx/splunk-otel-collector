#include "splunk.h"
#include "config.h"
#include "metrics_client.h"
#include "args.h"

#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#define ENV_VAR_LEN 512
#define MAX_CONFIG_ATTR_LEN 256
#define MAX_CMDLINE_LEN 16000
#define MAX_ARGS 256

#define JAVA_TOOL_OPTIONS_PREFIX "-javaagent:";

static char *const conf_file = "/usr/lib/splunk-instrumentation/instrumentation.conf";

extern char *program_invocation_short_name;

bool has_read_access(const char *s);

void set_java_tool_options(logger log, struct config *cfg);

void get_service_name_from_cmdline(logger log, char *dest, cmdline_reader cr);

bool is_disable_env_set();

bool is_java_tool_options_set();

void set_env_var(logger log, const char *var_name, const char *value);

void set_env_var_from_attr(logger log, const char *attr_name, const char *env_var_name, const char *value);

// The entry point for all executables prior to their execution. If the executable is named "java", we
// set the env vars JAVA_TOOL_OPTIONS to the path of the java agent jar and OTEL_SERVICE_NAME to the
// service name based on the arguments to the java command.
void __attribute__((constructor)) splunk_instrumentation_enter() {
    logger l = new_logger();
    cmdline_reader cr = new_cmdline_reader();
    if (cr == NULL) {
        return;
    }
    auto_instrument(l, has_read_access, program_invocation_short_name, load_config, cr, send_otlp_metric);
    cmdline_reader_close(cr);
    free_logger(l);
}

void auto_instrument(
        logger log,
        has_access_func_t has_access,
        const char *program_name,
        load_config_func_t load_config_func,
        cmdline_reader cr,
        send_otlp_metric_func_t send_otlp_metric_func
) {
    if (!streq(program_name, "java")) {
        return;
    }
    if (is_disable_env_set()) {
        log_debug(log, "disable_env set, quitting");
        return;
    }
    if (is_java_tool_options_set()) {
        log_debug(log, "java_tool_options set, quitting");
        return;
    }

    struct config cfg = {
            .java_agent_jar = NULL,
            .resource_attributes = NULL,
            .service_name = NULL,
            .disable_telemetry = NULL
    };
    load_config_func(log, &cfg, conf_file);
    if (cfg.java_agent_jar == NULL) {
        log_warning(log, "java_agent_jar not set, quitting");
        return;
    }

    if (!has_access(cfg.java_agent_jar)) {
        log_info(log, "agent jar not found or no read access, quitting");
        return;
    }

    char service_name[MAX_CMDLINE_LEN] = "";
    if (cfg.service_name == NULL) {
        get_service_name_from_cmdline(log, service_name, cr);
    } else {
        strncpy(service_name, cfg.service_name, MAX_CMDLINE_LEN);
    }

    if (strlen(service_name) == 0) {
        log_info(log, "service name empty, quitting");
        return;
    }

    set_env_var(log, otel_service_name_var, service_name);

    set_java_tool_options(log, &cfg);

    set_env_var_from_attr(log, "resource_attributes", resource_attributes_var, cfg.resource_attributes);

    if (eq_true(cfg.disable_telemetry)) {
        log_info(log, "disabling telemetry as per config");
    } else {
        send_otlp_metric_func(log, service_name);
    }

    free_config(&cfg);
}

void get_service_name_from_cmdline(logger log, char *dest, cmdline_reader cr) {
    char *args[MAX_ARGS];
    int n = get_cmdline_args(args, cr, MAX_ARGS, MAX_CMDLINE_LEN, log);
    generate_servicename_from_args(dest, args, n);
    free_cmdline_args(args, n);
}

void set_env_var_from_attr(logger log, const char *attr_name, const char *env_var_name, const char *value) {
    if (value == NULL) {
        return;
    }
    size_t len = strlen(value);
    if (len > MAX_CONFIG_ATTR_LEN) {
        char log_line[MAX_LOG_LINE_LEN] = "";
        sprintf(log_line, "%s too long: got %zu chars, max %d chars", attr_name, len, MAX_CONFIG_ATTR_LEN);
        log_warning(log, log_line);
        return;
    }
    set_env_var(log, env_var_name, value);
}

void set_env_var(logger log, const char *var_name, const char *value) {
    char log_line[MAX_LOG_LINE_LEN] = "";
    sprintf(log_line, "setting %s='%s'", var_name, value);
    log_debug(log, log_line);
    setenv(var_name, value, 0);
}

void set_java_tool_options(logger log, struct config *cfg) {
    char java_tool_options[ENV_VAR_LEN] = JAVA_TOOL_OPTIONS_PREFIX;
    char log_line[MAX_LOG_LINE_LEN] = "";
    size_t jar_path_len = strlen(cfg->java_agent_jar);
    if (jar_path_len > MAX_CONFIG_ATTR_LEN) {
        sprintf(log_line, "jar_path too long: got %zu chars, max %d chars", jar_path_len, MAX_CONFIG_ATTR_LEN);
        log_warning(log, log_line);
        return;
    }
    strcat(java_tool_options, (*cfg).java_agent_jar);
    sprintf(log_line, "setting JAVA_TOOL_OPTIONS='%s'", java_tool_options);
    log_debug(log, log_line);
    setenv(java_tool_options_var, java_tool_options, 0);
}

bool is_disable_env_set() {
    char *env = getenv(disable_env_var);
    return env && !streq("false", env) && !streq("FALSE", env) && !streq("0", env);
}

bool is_java_tool_options_set() {
    char *env = getenv(java_tool_options_var);
    return env != NULL && strlen(env) > 0;
}

bool has_read_access(const char *s) {
    return access(s, R_OK) == 0;
}

bool streq(const char *expected, const char *actual) {
    if (expected == NULL && actual == NULL) {
        return true;
    }
    if (expected == NULL || actual == NULL) {
        return false;
    }
    return strcmp(expected, actual) == 0;
}
