#include <stdio.h>
#include <stdlib.h>
#include "test_utils.h"
#include "splunk.h"

void require_true(char *funcname, bool actual) {
    if (!actual) {
        printf("%s: require_true: got false\n", funcname);
        fail();
    }
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

void require_env(char *funcname, char *expected, char *env_var) {
    char *env = getenv(env_var);
    if (!streq(expected, env)) {
        printf("%s: require_env: %s: expected [%s] got [%s]\n", funcname, env_var, expected, env);
        fail();
    }
}

void require_unset_env(char *funcname, char *env_var) {
    char *env = getenv(env_var);
    if (env) {
        printf("%s: require_unset_env: %s: expected unset got [%s]\n", funcname, env_var, env);
        fail();
    }
}

void fail() {
    exit(EXIT_FAILURE);
}
