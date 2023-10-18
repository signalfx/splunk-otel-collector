#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#define MAX_LINE_LENGTH 1023
#define MAX_LINES 50

#define ALLOWED_ENV_VARS "OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_RESOURCE_ATTRIBUTES", "SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED", "JAVA_TOOL_OPTIONS", "NODE_OPTIONS"

static char *const allowed_env_vars[] = {ALLOWED_ENV_VARS};
static size_t const allowed_env_vars_size = sizeof(allowed_env_vars) / sizeof(*allowed_env_vars);

#define JAVA_ENV_VAR_FILE "/etc/splunk/zeroconfig/java.conf"
#define NODEJS_ENV_VAR_FILE "/etc/splunk/zeroconfig/node.conf"

// TODO change to systemd drop in file paths
static char *const env_var_file_java = JAVA_ENV_VAR_FILE;
static char *const env_var_file_node = NODEJS_ENV_VAR_FILE;

extern char *program_invocation_short_name;

// The entry point for all executables prior to their execution.
void __attribute__((constructor)) enter() {
    char *env_var_file;
    if (strcmp("java", program_invocation_short_name) == 0) {
        env_var_file = env_var_file_java;
    } else if (strcmp("node", program_invocation_short_name) == 0) {
        env_var_file = env_var_file_node;
    } else {
        // we don't want to inject environment variables for this program.
        return;
    }

    if (MAX_LINES <= 0 || MAX_LINE_LENGTH <= 0) {
        return;
    }

    const size_t buffer_size = MAX_LINE_LENGTH + 1;
    char buffer[buffer_size];

    FILE *fp = fopen(env_var_file, "r");
    if (fp == NULL) {
        return;
    }

    int line_count = 0;

    while (fgets(buffer, buffer_size, fp) != NULL) {
        line_count += 1;
        if (line_count > MAX_LINES) {
            break;
        }

        char *newline = memchr(buffer, '\n', buffer_size);
        if (newline != NULL) {
            // terminate the string inside buffer at the newline
            *newline = '\0';
        }
        // if we have read a string without a newline termination, we have reached the end of file
        // or the string is past our max buffer size and we have an invalid value and we need to abort.

        // Properly formatted input file contains lines of length less than MAX_LINE_LENGTH and ends with a newline.
        // If newline character is not read, it means that a line is greater that MAX_LINE_LENGTH or the file does not end with a newline.
        if (newline == NULL) {
            break;
        }
        if (strnlen(buffer, buffer_size) == 0) {
            continue;
        }
        if (buffer[0] == '#') {
    		continue;
    	}

        char *equals = memchr(buffer, '=', buffer_size);
        if (equals == NULL) {
            continue;
        }

        // buffer is key=value\0

        *equals = '\0';

        // buffer is key\0value\0

        char *key = buffer;
        char *value = equals + 1;

        // check if key is allowed:
        if (strchr(key, ' ') != NULL) {
            continue;
        }

        // check if key is allowed
        for (int i = 0; i < allowed_env_vars_size; i++) {
            if (strcmp(allowed_env_vars[i], key) == 0) {
                setenv(key, value, 0);
                break;
            }
        }
    }
    fclose(fp);
}
