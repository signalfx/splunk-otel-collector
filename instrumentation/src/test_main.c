#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "splunk.h"
#include "config.h"
#include "test_main.h"

#define NUM_TESTS 9

int main(void) {
    run_tests();
    puts("PASS");
    return EXIT_SUCCESS;
}

void run_tests() {
    test_func_t *tests[NUM_TESTS] = {
            test_auto_instrument_success,
            test_auto_instrument_not_java,
            test_auto_instrument_no_access,
            test_auto_instrument_splunk_env_var_true,
            test_auto_instrument_splunk_env_var_false,
            test_auto_instrument_splunk_env_var_false_caps,
            test_auto_instrument_splunk_env_var_zero,
            test_read_config,
            test_read_config_missing_file
    };
    for (int i = 0; i < NUM_TESTS; ++i) {
        run_test(tests[i]);
    }
}

void run_test(test_func_t f) {
    unsetenv(JAVA_TOOL_OPTIONS);
    unsetenv(disable_env_var_name);
    logger l = new_logger();
    f(l);
    free_logger(l);
}

void test_auto_instrument_success(logger l) {
    auto_instrument(l, access_check_true, "java", fake_load_config);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_auto_instrument_success";
    require_equal_ints(funcname, 2, n);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[0]);
    require_equal_strings(funcname, "setting OTEL_RESOURCE_ATTRIBUTES='service.name=my.service'", logs[1]);
    require_equal_strings(funcname, "-javaagent:/foo/asdf.jar", getenv("JAVA_TOOL_OPTIONS"));
    require_equal_strings(funcname, "service.name=my.service", getenv("OTEL_RESOURCE_ATTRIBUTES"));
}

void test_auto_instrument_not_java(logger l) {
    auto_instrument(l, access_check_true, "foo", fake_load_config);
    char *env = getenv(JAVA_TOOL_OPTIONS);
    if (env) {
        fail();
    }
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_ints("test_auto_instrument_not_java", 0, n);
}

void test_auto_instrument_no_access(logger l) {
    auto_instrument(l, access_check_false, "java", fake_load_config);
    char *env = getenv(JAVA_TOOL_OPTIONS);
    if (env) {
        exit(EXIT_FAILURE);
    }
    char *logs[256];
    char *funcname = "test_auto_instrument_no_access";
    require_equal_ints(funcname, 1, get_logs(l, logs));
    require_equal_strings(funcname, "agent jar not found or no read access, quitting", logs[0]);
}

void test_auto_instrument_splunk_env_var_true(logger l) {
    setenv(disable_env_var_name, "true", 0);
    auto_instrument(l, access_check_true, "java", fake_load_config);
    require_unset_env("test_auto_instrument_splunk_env_var_true", JAVA_TOOL_OPTIONS);
}

void test_auto_instrument_splunk_env_var_false(logger l) {
    setenv(disable_env_var_name, "false", 0);
    auto_instrument(l, access_check_true, "java", fake_load_config);
    require_env("test_auto_instrument_splunk_env_var_false", JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

void test_auto_instrument_splunk_env_var_false_caps(logger l) {
    setenv(disable_env_var_name, "FALSE", 0);
    auto_instrument(l, access_check_true, "java", fake_load_config);
    require_env("test_auto_instrument_splunk_env_var_false_caps", JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

void test_auto_instrument_splunk_env_var_zero(logger l) {
    setenv(disable_env_var_name, "0", 0);
    auto_instrument(l, access_check_true, "java", fake_load_config);
    require_env("test_auto_instrument_splunk_env_var_zero", JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

void test_read_config(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, "config.example.txt");
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config";
    require_equal_ints(funcname, 1, n);
    require_equal_strings(funcname, "reading config file: config.example.txt", logs[0]);
    require_equal_strings(funcname, "my.service", cfg.service_name);
    require_equal_strings(funcname, "/foo/bar/baz.jar", cfg.java_agent_jar);
    free_config(&cfg);
}

void test_read_config_missing_file(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, "foo.txt");
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config_missing_file";
    require_equal_ints(funcname, 3, n);
    require_equal_strings(funcname, "file not found: foo.txt", logs[0]);
    require_equal_strings(funcname, "service_name not found in config, using default: default.service", logs[1]);
    require_equal_strings(funcname, "java_agent_jar not found in config, using default: /usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", logs[2]);
    require_equal_strings(funcname, "default.service", cfg.service_name);
    require_equal_strings(funcname, "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", cfg.java_agent_jar);
    free_config(&cfg);
}

void fake_load_config(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->service_name = strdup("my.service");
}

bool access_check_true(const char *s) {
    return true;
}

bool access_check_false(const char *s) {
    return false;
}

void require_equal_strings(char *funcname, char *expected, char *actual) {
    if (!streq(expected, actual)) {
        printf("%s: require_equal_strings: expected [%s] got [%s]\n", funcname, expected, actual);
        fail();
    }
}

void require_equal_ints(char *funcname, int expected, int actual) {
    if (expected != actual) {
        printf("%s: require_equal_ints: expected [%d] got [%d]\n", funcname, expected, actual);
        fail();
    }
}

void require_env(char *funcname, char *env_var, char *expected) {
    char *env = getenv(env_var);
    if (!streq(expected, env)) {
        printf("%s: require_env: %s: expected [%s] got [%s]\n", funcname, env_var, expected, env);
        fail();
    }
}

void require_unset_env(char *funcname, char *env_var) {
    char *env = getenv(env_var);
    if (env) {
        printf("%s: require_unset_env: %s: expected unset got [%s]", funcname, env_var, env);
        fail();
    }
}

void fail() {
    exit(EXIT_FAILURE);
}
