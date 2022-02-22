#ifndef SPLUNK_INSTRUMENTATION_TEST_MAIN_H
#define SPLUNK_INSTRUMENTATION_TEST_MAIN_H

#include "logger.h"
#include "config.h"
#include "cmdline_reader.h"
#include <stdbool.h>

typedef void (test_func_t)(logger);

void run_tests();

void run_test(test_func_t run_test);

void test_auto_instrument_not_java(logger l);

void test_auto_instrument_svc_name_in_config(logger l);

void test_auto_instrument_no_svc_name_in_config(logger l);

void test_auto_instrument_no_access(logger l);

void test_auto_instrument_splunk_env_var_true(logger l);

void test_auto_instrument_splunk_env_var_false(logger l);

void test_auto_instrument_splunk_env_var_false_caps(logger l);

void test_auto_instrument_splunk_env_var_zero(logger l);

void test_read_config(logger l);

void test_read_config_missing_file(logger l);

void test_read_args_simple(logger l);

void test_read_args_limit(logger l);

void test_extract_servicename_from_args_tomcat(logger l);

void test_extract_servicename_from_args_simple_jar(logger l);

void test_extract_servicename_from_args_module(logger l);

void test_clean_jars(logger l);

void test_clean_jar(logger l);

void test_is_unique_path_element(logger l);

void test_truncate_jar(logger l);

void test_truncate_jar_short(logger l);

void test_dots_to_dashes(logger l);

void test_env_var_already_set(logger l);

// fakes/testdata

void fake_load_config(logger log, struct config *cfg, char *path);

void fake_load_config_no_svcname(logger log, struct config *cfg, char *path);

cmdline_reader new_default_test_cmdline_reader();

bool access_check_true(const char *s);

bool access_check_false(const char *s);

int tomcat_args(char *args[]);

int petclinic_args(char *args[]);

int module_args(char *args[]);

#endif //SPLUNK_INSTRUMENTATION_TEST_MAIN_H
