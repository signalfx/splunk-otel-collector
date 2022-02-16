#ifndef SPLUNK_INSTRUMENTATION_TEST_MAIN_H
#define SPLUNK_INSTRUMENTATION_TEST_MAIN_H

#include <stdbool.h>
#include "logger.h"
#include "config.h"

static char *const JAVA_TOOL_OPTIONS = "JAVA_TOOL_OPTIONS";

typedef void (test_func_t)(logger);

void fake_load_config(logger log, struct config *cfg, char *path);

void fail();

void run_tests();

void run_test(test_func_t f);

bool access_check_true(const char *s);

bool access_check_false(const char *s);

void test_auto_instrument_not_java(logger l);

void test_auto_instrument_success(logger l);

void test_auto_instrument_no_access(logger l);

void test_auto_instrument_splunk_env_var_true(logger l);

void test_auto_instrument_splunk_env_var_false(logger l);

void test_auto_instrument_splunk_env_var_false_caps(logger l);

void test_auto_instrument_splunk_env_var_zero(logger l);

void test_read_config(logger l);

void test_read_config_missing_file(logger l);

void require_equal_ints(char *funcname, int expected, int actual);

void require_equal_strings(char *funcname, char *expected, char *actual);

void require_env(char *funcname, char *env_var, char *expected);

void require_unset_env(char *funcname, char *env_var);

#endif //SPLUNK_INSTRUMENTATION_TEST_MAIN_H
