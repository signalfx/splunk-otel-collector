#include "splunk.h"
#include "config.h"
#include "metrics_client.h"
#include "args.h"

#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#define MAX_ENV_VAR_LEN 512
#define MAX_CONFIG_ATTR_LEN 256
#define MAX_CMDLINE_LEN 16000
#define MAX_ARGS 256

#define JAVA_TOOL_OPTIONS_PREFIX "-javaagent:";

static char *const conf_file = "/usr/lib/splunk-instrumentation/instrumentation.conf";
static char *const prof_enabled_cmdline_switch = " -Dsplunk.profiler.enabled=true";
static char *const prof_memory_enabled_cmdline_switch = " -Dsplunk.profiler.memory.enabled=true";
static char *const metrics_enabled_cmdline_switch = " -Dsplunk.metrics.enabled=true";

extern char *program_invocation_short_name;

int has_read_access(const char *s);

void set_java_tool_options(logger log, struct config *cfg);

void get_service_name_from_cmdline(logger log, char *dest, cmdline_reader cr);

int is_disable_env_set();

int is_java_tool_options_set();

void set_env_var(logger log, const char *var_name, const char *value);

void set_env_var_from_attr(logger log, const char *attr_name, const char *env_var_name, const char *value);

void get_service_name(logger log, cmdline_reader cr, struct config *cfg, char *dest);

void log_line_length_warning(logger log, char *log_line);

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
            .disable_telemetry = NULL,
            .generate_service_name = NULL,
            .enable_profiler = NULL,
            .enable_profiler_memory = NULL,
            .enable_metrics = NULL
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
    if (str_to_bool(cfg.generate_service_name, 1)) {
        get_service_name(log, cr, &cfg, service_name);
        if (strlen(service_name) == 0) {
            log_info(log, "service name empty, quitting");
            return;
        }
        set_env_var(log, otel_service_name_var, service_name);
    } else {
        log_debug(log, "service name generation explicitly disabled");
    }

    set_java_tool_options(log, &cfg);

    set_env_var_from_attr(log, "resource_attributes", resource_attributes_var, cfg.resource_attributes);

    if (str_to_bool(cfg.disable_telemetry, 0)) {
        log_info(log, "disabling telemetry as per config");
    } else {
        send_otlp_metric_func(log, service_name);
    }

    free_config(&cfg);
}

void get_service_name(logger log, cmdline_reader cr, struct config *cfg, char *dest) {
    if (cfg->service_name == NULL) {
        get_service_name_from_cmdline(log, dest, cr);
    } else {
        strncpy(dest, (*cfg).service_name, MAX_CMDLINE_LEN);
    }
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
    char java_tool_options[MAX_ENV_VAR_LEN] = JAVA_TOOL_OPTIONS_PREFIX;
    char log_line[MAX_LOG_LINE_LEN] = "";
    size_t jar_path_len = strlen(cfg->java_agent_jar);
    if (jar_path_len > MAX_CONFIG_ATTR_LEN) {
        sprintf(log_line, "jar_path too long: got %zu chars, max %d chars", jar_path_len, MAX_CONFIG_ATTR_LEN);
        log_warning(log, log_line);
        return;
    }
    int remaining = concat_strings(java_tool_options, cfg->java_agent_jar, MAX_ENV_VAR_LEN);
    // It's not possible (at time of writing) for `remaining` to be less than zero in this function because the sum of
    // MAX_CONFIG_ATTR_LEN (256) and the lengths of the pre-defined command line switches will always be less than
    // MAX_ENV_VAR_LEN (512), but check for truncation anyway to account for future additions.
    if (remaining < 0) {
        log_line_length_warning(log, log_line);
        return;
    }
    if (str_to_bool(cfg->enable_profiler, 0)) {
        remaining = concat_strings(java_tool_options, prof_enabled_cmdline_switch, MAX_ENV_VAR_LEN);
        if (remaining < 0) {
            log_line_length_warning(log, log_line);
            return;
        }
    }
    if (str_to_bool(cfg->enable_profiler_memory, 0)) {
        remaining = concat_strings(java_tool_options, prof_memory_enabled_cmdline_switch, MAX_ENV_VAR_LEN);
        if (remaining < 0) {
            log_line_length_warning(log, log_line);
            return;
        }
    }
    if (str_to_bool(cfg->enable_metrics, 0)) {
        remaining = concat_strings(java_tool_options, metrics_enabled_cmdline_switch, MAX_ENV_VAR_LEN);
        if (remaining < 0) {
            log_line_length_warning(log, log_line);
            return;
        }
    }
    sprintf(log_line, "setting JAVA_TOOL_OPTIONS='%s'", java_tool_options);
    log_debug(log, log_line);
    setenv(java_tool_options_var, java_tool_options, 0);
}

void log_line_length_warning(logger log, char *log_line) {
    sprintf(log_line, "excessive line length: not setting JAVA_TOOL_OPTIONS");
    log_warning(log, log_line);
}

// concat_strings concatenates the string defined by src to the memory location pointed to by dest,
// returning the number of characters remaining in the dest buffer. The return value may be negative
// indicating the number of characters that were not concatenated because of a lack of space. The tot_dest_size
// argument indicates the total number of bytes in the dest memory array. See unit tests for examples.
int concat_strings(char *dest, char *src, int tot_dest_size) {
    int orig_dest_len = (int) strlen(dest);
    strncat(dest, src, tot_dest_size - 1);
    return tot_dest_size - (int) orig_dest_len - (int) strlen(src) - 1;
}

int is_disable_env_set() {
    char *env = getenv(disable_env_var);
    return env && !streq("false", env) && !streq("FALSE", env) && !streq("0", env);
}

int is_java_tool_options_set() {
    char *env = getenv(java_tool_options_var);
    return env != NULL && strlen(env) > 0;
}

int has_read_access(const char *s) {
    return access(s, R_OK) == 0;
}

int streq(const char *expected, const char *actual) {
    if (expected == NULL && actual == NULL) {
        return 1;
    }
    if (expected == NULL || actual == NULL) {
        return 0;
    }
    return strcmp(expected, actual) == 0;
}
