#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>

#define LINE_BUFFER_LENGTH 1024

static char *const env_var_file = "/etc/splunk/zeroconfig.conf";

static int allowed_env_vars_size = 2;

static char *const allowed_env_vars[] = {
    "OTEL_SERVICE_NAME",
    "OTEL_RESOURCE_ATTRIBUTES",
};


// The entry point for all executables prior to their execution.
void __attribute__((constructor)) splunk_instrumentation_enter() {
    FILE * fp;
    char buffer[LINE_BUFFER_LENGTH];

    fp = fopen(env_var_file, "r");
    if (fp == NULL)
        return;

    while(fgets(buffer, LINE_BUFFER_LENGTH, fp) != NULL) {
        buffer[strcspn(buffer, "\n")] = '\0';
        if (strlen(buffer) == 0) {
            continue;
        }
        if (buffer[0] == '#') {
    		continue;
    	}

        char *value = strchr(buffer, '=');
        if (value == NULL) {
            continue;
        }
        int index = (int)(value - &buffer[0]);
        value++;

        char key[index];
        memcpy(key, &buffer[0], index);
        key[index] = '\0';

        for (int i = 0; i < allowed_env_vars_size; i++) {
            if (strcmp(allowed_env_vars[i], key) == 0) {
                setenv(key, value, 0);
                break;
            }
        }
    }
    fclose(fp);
}
