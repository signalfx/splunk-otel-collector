#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "splunk.h"

#define FLAG_LEN 256

#define FLAG_PREFIX "-javaagent:";

extern char *program_invocation_short_name;
char *const AGENT_JAR_PATH = "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar";

// The entry point for all executables prior to their execution. If the executable is named "java", we
// set JAVA_TOOL_OPTIONS to the path of the java agent jar.
void __attribute__((constructor)) splunk_instrumentation_enter() {
    logger l = new_logger();
    auto_instrument(l, has_read_access, program_invocation_short_name, AGENT_JAR_PATH);
    free_logger(l);
}

void auto_instrument(logger log, has_access_t has_access, const char *program_name, const char *jar_path) {
    if (strcmp(program_name, "java") != 0) {
        return;
    }
    if (is_disable_env_set()) {
        log_debug(log, "disable_env_set, quitting");
        return;
    }
    if (!has_access(jar_path)) {
        log_info(log, "agent jar not found or no read access, quitting");
        return;
    }
    char opts[FLAG_LEN] = FLAG_PREFIX;
    if (strlen(opts) + strlen(jar_path) > FLAG_LEN) {
        log_info(log, "jar path too long");
        return;
    }
    strncat(opts, jar_path, FLAG_LEN);
    log_debug(log, "setting JAVA_TOOL_OPTIONS");
    setenv("JAVA_TOOL_OPTIONS", opts, 0);
}

bool streq(char *a, char *b) {
    return strcmp(a, b) == 0;
}

bool is_disable_env_set() {
    char *env = getenv(DISABLE_ENV_VAR_NAME);
    return env && !streq("false", env) && !streq("FALSE", env) && !streq("0", env);
}

bool has_read_access(const char *s) {
    return access(s, R_OK) == 0;
}
