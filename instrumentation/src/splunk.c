#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#include "config.h"
#include "splunk.h"

#define ENV_VAR_LEN 512
#define MAX_SERVICE_NAME_LEN 256
#define MAX_JAR_PATH_LEN 256

#define JAVA_TOOL_OPTIONS_PREFIX "-javaagent:";
#define OTEL_RESOURCE_ATTRIBUTES_PREFIX "service.name="

extern char *program_invocation_short_name;

bool has_read_access(const char *s);

void set_java_tool_options(logger log, struct config *cfg);

void set_service_name(logger log, struct config *cfg);

bool is_disable_env_set();

// The entry point for all executables prior to their execution. If the executable is named "java", we
// set JAVA_TOOL_OPTIONS to the path of the java agent jar.
void __attribute__((constructor)) splunk_instrumentation_enter() {
    logger l = new_logger();
    auto_instrument(l, has_read_access, program_invocation_short_name, load_config);
    free_logger(l);
}

void auto_instrument(logger log, has_access_func_t has_access, const char *program_name, load_config_func_t load_config_func) {
    if (!streq(program_name, "java")) {
        return;
    }
    if (is_disable_env_set()) {
        log_debug(log, "disable_env_set, quitting");
        return;
    }

    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config_func(log, &cfg, conf_file);

    if (!has_access(cfg.java_agent_jar)) {
        log_info(log, "agent jar not found or no read access, quitting");
        return;
    }

    set_java_tool_options(log, &cfg);
    set_service_name(log, &cfg);

    free_config(&cfg);
}

void set_service_name(logger log, struct config *cfg) {
    char otel_resource_attributes[ENV_VAR_LEN] = OTEL_RESOURCE_ATTRIBUTES_PREFIX;
    char log_line[MAX_LOG_LINE_LEN];
    size_t service_name_len = strlen(((*cfg).service_name));
    if (service_name_len > MAX_SERVICE_NAME_LEN) {
        sprintf(log_line, "service_name too long: got %zu chars, max %d chars", service_name_len, MAX_SERVICE_NAME_LEN);
        log_warning(log, log_line);
        return;
    }
    strcat(otel_resource_attributes, (*cfg).service_name);
    sprintf(log_line, "setting OTEL_RESOURCE_ATTRIBUTES='%s'", otel_resource_attributes);
    log_debug(log, log_line);
    setenv("OTEL_RESOURCE_ATTRIBUTES", otel_resource_attributes, 0);
}

void set_java_tool_options(logger log, struct config *cfg) {
    char java_tool_options[ENV_VAR_LEN] = JAVA_TOOL_OPTIONS_PREFIX;
    char log_line[MAX_LOG_LINE_LEN];
    size_t jar_path_len = strlen(cfg->java_agent_jar);
    if (jar_path_len > MAX_JAR_PATH_LEN) {
        sprintf(log_line, "jar_path too long: got %zu chars, max %d chars", jar_path_len, MAX_JAR_PATH_LEN);
        log_warning(log, log_line);
        return;
    }
    strcat(java_tool_options, (*cfg).java_agent_jar);
    sprintf(log_line, "setting JAVA_TOOL_OPTIONS='%s'", java_tool_options);
    log_debug(log, log_line);
    setenv("JAVA_TOOL_OPTIONS", java_tool_options, 0);
}

bool is_disable_env_set() {
    char *env = getenv(disable_env_var_name);
    return env && !streq("false", env) && !streq("FALSE", env) && !streq("0", env);
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
