#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "splunk.h"
#include "test_main.h"

#define NUM_TESTS 7

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
            test_auto_instrument_splunk_env_var_zero
    };
    for (int i = 0; i < NUM_TESTS; ++i) {
        run_test(tests[i]);
    }
}

void run_test(test_func_t f) {
    unsetenv(JAVA_TOOL_OPTIONS);
    unsetenv(DISABLE_ENV_VAR_NAME);
    logger l = new_logger();
    f(l);
    free_logger(l);
}

void test_auto_instrument_success(logger l) {
    auto_instrument(l, access_check_true, "java", "/foo/xyz.jar");
    char *logs[TEST_LOGS_MAX_LEN];
    int n = get_logs(l, logs);
    require_equal_ints(1, n);
    require_equal_strings("-javaagent:/foo/xyz.jar", getenv(JAVA_TOOL_OPTIONS));
}

void test_auto_instrument_not_java(logger l) {
    auto_instrument(l, access_check_true, "foo", "/bar/xyz.jar");
    char *env = getenv(JAVA_TOOL_OPTIONS);
    if (env) {
        fail();
    }
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_ints(0, n);
}

void test_auto_instrument_no_access(logger l) {
    auto_instrument(l, access_check_false, "java", "/bar/xyz.jar");
    char *env = getenv(JAVA_TOOL_OPTIONS);
    if (env) {
        exit(EXIT_FAILURE);
    }
    char *logs[256];
    require_equal_ints(1, get_logs(l, logs));
    require_equal_strings("agent jar not found or no read access, quitting", logs[0]);
}

void test_auto_instrument_splunk_env_var_true(logger l) {
    setenv(DISABLE_ENV_VAR_NAME, "true", 0);
    auto_instrument(l, access_check_true, "java", "/foo/asdf.jar");
    require_unset_env(JAVA_TOOL_OPTIONS);
}

void test_auto_instrument_splunk_env_var_false(logger l) {
    setenv(DISABLE_ENV_VAR_NAME, "false", 0);
    auto_instrument(l, access_check_true, "java", "/foo/asdf.jar");
    require_env(JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

void test_auto_instrument_splunk_env_var_false_caps(logger l) {
    setenv(DISABLE_ENV_VAR_NAME, "FALSE", 0);
    auto_instrument(l, access_check_true, "java", "/foo/asdf.jar");
    require_env(JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

void test_auto_instrument_splunk_env_var_zero(logger l) {
    setenv(DISABLE_ENV_VAR_NAME, "0", 0);
    auto_instrument(l, access_check_true, "java", "/foo/asdf.jar");
    require_env(JAVA_TOOL_OPTIONS, "-javaagent:/foo/asdf.jar");
}

bool access_check_true(const char *s) {
    return true;
}

bool access_check_false(const char *s) {
    return false;
}

void require_equal_strings(char *expected, char *actual) {
    if (strcmp(expected, actual) != 0) {
        printf("require_equal_strings: expected [%s] got [%s]\n", expected, actual);
        fail();
    }
}

void require_equal_ints(int expected, int actual) {
    if (expected != actual) {
        printf("require_equal_ints: expected [%d] got [%d]\n", expected, actual);
        fail();
    }
}

void require_env(char *env_var, char *expected) {
    char *env = getenv(env_var);
    if (env == NULL || strcmp(expected, env) != 0) {
        printf("require_env: %s: expected [%s] got [%s]\n", env_var, expected, env);
        fail();
    }
}

void require_unset_env(char *env_var) {
    char *env = getenv(env_var);
    if (env) {
        printf("require_unset_env: %s: expected unset got [%s]", env_var, env);
        fail();
    }
}

void fail() {
    exit(EXIT_FAILURE);
}
