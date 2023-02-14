#ifndef INSTRUMENTATION_TEST_UTILS_H
#define INSTRUMENTATION_TEST_UTILS_H

void print_logs(char **logs, int n);

void require_true(char *funcname, int actual);

void require_false(char *funcname, int actual);

void require_equal_strings(char *funcname, char *expected, char *actual);

void require_equal_ints(char *funcname, int expected, int actual);

void require_env(char *funcname, char *expected, char *env_var);

void require_env_len(char *funcname, int expected_len, char *env_var);

void require_unset_env(char *funcname, char *env_var);

void fail();

#endif //INSTRUMENTATION_TEST_UTILS_H
