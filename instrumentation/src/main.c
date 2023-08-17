#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#define MAX_LINE_LENGTH 1023
#define MAX_LINES 50

static char *const env_var_file = "/etc/splunk/zeroconfig.conf";

static char *const allowed_env_vars[] = {
    "OTEL_SERVICE_NAME",
    "OTEL_EXPORTER_OTLP_ENDPOINT",
    "OTEL_RESOURCE_ATTRIBUTES",
    "SPLUNK_PROFILER_ENABLED",
    "SPLUNK_PROFILER_MEMORY_ENABLED",
    "SPLUNK_METRICS_ENABLED",
};

const size_t allowed_env_vars_size = sizeof(allowed_env_vars) / sizeof(*allowed_env_vars);


// The entry point for all executables prior to their execution.
void __attribute__((constructor)) enter() {
    const size_t buffer_size = MAX_LINE_LENGTH + 1;
    char buffer[buffer_size];

    if (MAX_LINES <= 0 || MAX_LINE_LENGTH <= 0) {
        return;
    }

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

        char *newline = strchr(buffer, '\n');
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
        if (strlen(buffer) == 0) {
            continue;
        }
        if (buffer[0] == '#') {
    		continue;
    	}

        char *equals = strchr(buffer, '=');
        if (equals == NULL) {
            continue;
        }

        // buffer is key=value\0

        *equals = '\0';

        // buffer is key\0value\0

        char *key = buffer;
        char *value = equals + 1;

        // check if key and value are valid environment key and value

        for (int i = 0; i < allowed_env_vars_size; i++) {
            if (strcmp(allowed_env_vars[i], key) == 0) {
                setenv(key, value, 0);
                break;
            }
        }
    }
    fclose(fp);
}
