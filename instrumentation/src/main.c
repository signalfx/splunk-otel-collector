#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdio.h>
#include <limits.h>

#define MAX_LINE_LENGTH 1023
#define MAX_LINES 50

#define ALLOWED_EXECUTABLE_PATHS_KEY "SPLUNK_INSTRUMENTATION_ALLOWED_EXECUTABLE_PATHS"
#define ALLOWED_WORKING_DIRS_KEY "SPLUNK_INSTRUMENTATION_ALLOWED_WORKING_DIRS"

#define ALLOWED_ENV_VARS "OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_RESOURCE_ATTRIBUTES", "SPLUNK_PROFILER_ENABLED", "SPLUNK_PROFILER_MEMORY_ENABLED", "SPLUNK_METRICS_ENABLED", "JAVA_TOOL_OPTIONS", "NODE_OPTIONS", "CORECLR_ENABLE_PROFILING", "CORECLR_PROFILER", "CORECLR_PROFILER_PATH", "DOTNET_ADDITIONAL_DEPS", "DOTNET_SHARED_STORE", "DOTNET_STARTUP_HOOKS", "OTEL_DOTNET_AUTO_HOME", "OTEL_DOTNET_AUTO_PLUGINS", "OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_METRICS_EXPORTER", "OTEL_LOGS_EXPORTER"

static char *const allowed_env_vars[] = {ALLOWED_ENV_VARS};
static size_t const allowed_env_vars_size = sizeof(allowed_env_vars) / sizeof(*allowed_env_vars);

#define DOTNET_ENV_VAR_FILE "/etc/splunk/zeroconfig/dotnet.conf"
#define JAVA_ENV_VAR_FILE "/etc/splunk/zeroconfig/java.conf"
#define NODEJS_ENV_VAR_FILE "/etc/splunk/zeroconfig/node.conf"

static char *const env_var_file_dotnet = DOTNET_ENV_VAR_FILE;
static char *const env_var_file_java = JAVA_ENV_VAR_FILE;
static char *const env_var_file_node = NODEJS_ENV_VAR_FILE;

extern char *program_invocation_short_name;

typedef enum {
    LINE_KV,
    LINE_SKIP,
    LINE_STOP
} line_result_t;

static int read_exe_path(char *buf, size_t size) {
    ssize_t len = readlink("/proc/self/exe", buf, size - 1);
    if (len < 0) {
        return -1;
    }
    if ((size_t)len >= size - 1) {
        return -1;
    }
    buf[len] = '\0';
    return 0;
}

static int read_cwd_path(char *buf, size_t size) {
    if (getcwd(buf, size) == NULL) {
        return -1;
    }
    return 0;
}

static int token_contains_dotdot(const char *token) {
    if (strcmp(token, "..") == 0) {
        return 1;
    }
    if (strstr(token, "/../") != NULL) {
        return 1;
    }
    size_t len = strnlen(token, PATH_MAX);
    if (len >= 3 && strcmp(token + len - 3, "/..") == 0) {
        return 1;
    }
    return 0;
}

static size_t strip_trailing_slashes(char *token, size_t token_size) {
    size_t len = strnlen(token, token_size);
    if (len == 0) {
        return 0;
    }
    if (len == 1 && token[0] == '/') {
        return 1;
    }
    while (len > 1 && token[len - 1] == '/') {
        token[len - 1] = '\0';
        len--;
    }
    return len;
}

static int path_prefix_match(const char *path, const char *token, size_t token_len) {
    if (token_len == 1 && token[0] == '/') {
        return path[0] == '/';
    }
    return strncmp(path, token, token_len) == 0 &&
           (path[token_len] == '\0' || path[token_len] == '/');
}

static int path_matches_any(const char *path, const char *list) {
    if (path == NULL || list == NULL || list[0] == '\0') {
        return 0;
    }

    const char *start = list;
    while (*start != '\0') {
        const char *colon = strchr(start, ':');
        size_t token_len;
        if (colon != NULL) {
            token_len = (size_t)(colon - start);
        } else {
            token_len = strnlen(start, PATH_MAX);
        }

        if (token_len == 0 || start[0] != '/') {
            if (colon == NULL) {
                break;
            }
            start = colon + 1;
            continue;
        }

        if (token_len >= PATH_MAX) {
            if (colon == NULL) {
                break;
            }
            start = colon + 1;
            continue;
        }

        char token_buf[PATH_MAX];
        memcpy(token_buf, start, token_len);
        token_buf[token_len] = '\0';

        if (token_contains_dotdot(token_buf)) {
            if (colon == NULL) {
                break;
            }
            start = colon + 1;
            continue;
        }

        char resolved[PATH_MAX];
        const char *match_token;
        if (realpath(token_buf, resolved) != NULL) {
            match_token = resolved;
        } else {
            match_token = token_buf;
        }

        char match_stripped[PATH_MAX];
        size_t match_len = strnlen(match_token, PATH_MAX);
        if (match_len >= PATH_MAX) {
            if (colon == NULL) {
                break;
            }
            start = colon + 1;
            continue;
        }
        memcpy(match_stripped, match_token, match_len + 1);
        size_t stripped_len = strip_trailing_slashes(match_stripped, PATH_MAX);
        if (stripped_len == 0) {
            if (colon == NULL) {
                break;
            }
            start = colon + 1;
            continue;
        }

        if (path_prefix_match(path, match_stripped, stripped_len)) {
            return 1;
        }

        if (colon == NULL) {
            break;
        }
        start = colon + 1;
    }
    return 0;
}

static line_result_t next_kv(FILE *fp, char *buffer, size_t buffer_size, int *line_count, char **key, char **value) {
    if (fgets(buffer, buffer_size, fp) == NULL) {
        return LINE_STOP;
    }

    (*line_count) += 1;
    if (*line_count > MAX_LINES) {
        return LINE_STOP;
    }

    char *newline = memchr(buffer, '\n', buffer_size);
    if (newline != NULL) {
        *newline = '\0';
    }
    if (newline == NULL) {
        return LINE_STOP;
    }
    if (strnlen(buffer, buffer_size) == 0) {
        return LINE_SKIP;
    }
    if (buffer[0] == '#') {
        return LINE_SKIP;
    }

    char *equals = memchr(buffer, '=', buffer_size);
    if (equals == NULL) {
        return LINE_SKIP;
    }

    *equals = '\0';
    *key = buffer;
    *value = equals + 1;

    if (strchr(*key, ' ') != NULL) {
        return LINE_SKIP;
    }

    return LINE_KV;
}

// The entry point for all executables prior to their execution.
void __attribute__((constructor)) enter() {
    char *env_var_file;
    if (strcmp("dotnet", program_invocation_short_name) == 0) {
        env_var_file = env_var_file_dotnet;
    } else if (strcmp("java", program_invocation_short_name) == 0) {
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

    char allowed_exec_paths[buffer_size];
    char allowed_working_dirs[buffer_size];
    int exec_paths_set = 0;
    int working_dirs_set = 0;

    int line_count = 0;
    char *key;
    char *value;
    line_result_t result;

    while (1) {
        result = next_kv(fp, buffer, buffer_size, &line_count, &key, &value);
        if (result == LINE_STOP) {
            break;
        }
        if (result == LINE_SKIP) {
            continue;
        }
        if (strcmp(key, ALLOWED_EXECUTABLE_PATHS_KEY) == 0) {
            snprintf(allowed_exec_paths, buffer_size, "%s", value);
            exec_paths_set = 1;
        } else if (strcmp(key, ALLOWED_WORKING_DIRS_KEY) == 0) {
            snprintf(allowed_working_dirs, buffer_size, "%s", value);
            working_dirs_set = 1;
        }
    }

    if (exec_paths_set) {
        char exe_path[PATH_MAX];
        if (read_exe_path(exe_path, sizeof(exe_path)) != 0 ||
            !path_matches_any(exe_path, allowed_exec_paths)) {
            fclose(fp);
            return;
        }
    }
    if (working_dirs_set) {
        char cwd[PATH_MAX];
        if (read_cwd_path(cwd, sizeof(cwd)) != 0 ||
            !path_matches_any(cwd, allowed_working_dirs)) {
            fclose(fp);
            return;
        }
    }

    rewind(fp);
    line_count = 0;

    while (1) {
        result = next_kv(fp, buffer, buffer_size, &line_count, &key, &value);
        if (result == LINE_STOP) {
            break;
        }
        if (result == LINE_SKIP) {
            continue;
        }

        for (size_t i = 0; i < allowed_env_vars_size; i++) {
            if (strcmp(allowed_env_vars[i], key) == 0) {
                setenv(key, value, 0);
                break;
            }
        }
    }
    fclose(fp);
}
