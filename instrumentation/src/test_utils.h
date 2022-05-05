#ifndef INSTRUMENTATION_TEST_UTILS_H
#define INSTRUMENTATION_TEST_UTILS_H

#include <stdbool.h>

void require_true(char *funcname, bool actual);

void require_false(char *funcname, bool actual);

void require_equal_strings(char *funcname, char *expected, char *actual);

void require_equal_ints(char *funcname, int expected, int actual);

void require_env(char *funcname, char *expected, char *env_var);

void require_unset_env(char *funcname, char *env_var);

void fail();

#endif //INSTRUMENTATION_TEST_UTILS_H
