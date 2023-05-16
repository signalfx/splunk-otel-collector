#ifndef SPLUNK_INSTRUMENTATION_TEST_MAIN_H
#define SPLUNK_INSTRUMENTATION_TEST_MAIN_H

#include "logger.h"
#include "config.h"
#include "cmdline_reader.h"

typedef void (test_func_t)(logger);

void run_tests();

void run_test(test_func_t run_test);

void test_auto_instrument_not_java(logger l);

void test_auto_instrument_svc_name_specified(logger l);

void test_auto_instrument_gen_svc_name_explicitly_enabled(logger l);

void test_auto_instrument_gen_svc_name_explicitly_disabled(logger l);

void test_auto_instrument_no_svc_name_in_config(logger l);

void test_auto_instrument_no_access(logger l);

void test_auto_instrument_splunk_env_var_true(logger l);

void test_auto_instrument_splunk_env_var_false(logger l);

void test_auto_instrument_splunk_env_var_false_caps(logger l);

void test_auto_instrument_splunk_env_var_zero(logger l);

void test_read_config_default(logger l);

void test_read_config_all_options(logger l);

void test_read_config_missing_file(logger l);

void test_read_args_simple(logger l);

void test_read_args_max_args_limit(logger l);

void test_read_args_max_cmdline_len_limit(logger l);

void test_extract_servicename_from_args_tomcat(logger l);

void test_extract_servicename_from_args_simple_jar(logger l);

void test_extract_servicename_from_args_module(logger l);

void test_extract_servicename_from_args_okhttp(logger l);

void test_extract_servicename_from_args_zk(logger l);

void test_extract_servicename_from_args_kafka(logger l);

void test_extract_servicename_from_args_spring(logger l);

void test_transform_multi_jars(logger l);

void test_tokenset(logger l);

void test_tokenset_overflow(logger l);

void test_transform_jar_path_elements(logger l);

void test_dedupe_hyphenated(logger l);

void test_is_unique_path_element(logger l);

void test_truncate_jar(logger l);

void test_truncate_jar_short(logger l);

void test_dots_to_dashes(logger l);

void test_env_var_already_set(logger l);

void test_is_legal_java_module_main_class(logger l);

void test_is_legal_java_fq_main_class(logger l);

void test_is_legal_java_package_element(logger l);

void test_is_legal_module(logger l);

void test_str_to_bool(logger l);

void test_enable_telemetry(logger l);

void test_disable_telemetry(logger l);

void test_enable_profiling(logger l);

void test_enable_profiling_memory(logger l);

void test_enable_metrics(logger l);

void test_concat_string_to_empty_just_enough_room(logger l);

void test_concat_string_to_empty_extra_room(logger l);

void test_concat_string_to_empty_not_enough_room(logger l);

void test_concat_string_to_nonempty_just_enough_room(logger l);

void test_long_cfg_attributes(logger l);

void test_auto_instrument_gen_svcname_disabled_but_specified(logger l);

// fakes/testdata

void fake_send_otlp_metric(logger log, char *service_name);

void fake_config_svcname_explicitly_specified(logger log, struct config *cfg, char *path);

void fake_config_generate_svcname_enabled(logger log, struct config *cfg, char *path);

void fake_config_generate_svcname_disabled(logger log, struct config *cfg, char *path);

void fake_config_generate_svcname_disabled_but_explicitly_specified(logger log, struct config *cfg, char *path);

void fake_config_no_svcname(logger log, struct config *cfg, char *path);

void fake_config_disable_telemetry_not_specified(logger log, struct config *cfg, char *path);

void fake_config_disable_telemetry_true(logger log, struct config *cfg, char *path);

void fake_config_enable_profiler(logger log, struct config *cfg, char *path);

void fake_config_enable_profiler_memory(logger log, struct config *cfg, char *path);

void fake_config_enable_metrics(logger log, struct config *cfg, char *path);

void fake_config_max_attributes(logger log, struct config *cfg, char *path);

cmdline_reader new_default_test_cmdline_reader();

int access_check_true(const char *s);

int access_check_false(const char *s);

int tomcat_args(char *args[]);

int petclinic_args(char *args[]);

int module_args(char *args[]);

int okhttp_and_jedis_args(char *args[]);

int zk_args(char *args[]);

int kafka_args(char *args[]);

int spring_args(char *args[]);

char *zk_classpath();

#endif //SPLUNK_INSTRUMENTATION_TEST_MAIN_H
