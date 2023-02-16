#include "test_main.h"
#include "splunk.h"
#include "config.h"
#include "args.h"
#include "test_utils.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char *const test_config_path = "testdata/instrumentation-default.conf";
static char *const test_config_path_all_options = "testdata/instrumentation-options.conf";

int main(void) {
    run_tests();
    puts("PASS");
    return EXIT_SUCCESS;
}

void run_tests() {
    test_func_t *tests[] = {
            test_auto_instrument_svc_name_specified,
            test_auto_instrument_gen_svc_name_explicitly_enabled,
            test_auto_instrument_gen_svc_name_explicitly_disabled,
            test_auto_instrument_no_svc_name_in_config,
            test_auto_instrument_not_java,
            test_auto_instrument_no_access,
            test_auto_instrument_splunk_env_var_true,
            test_auto_instrument_splunk_env_var_false,
            test_auto_instrument_splunk_env_var_false_caps,
            test_auto_instrument_splunk_env_var_zero,
            test_read_config_default,
            test_read_config_all_options,
            test_read_config_missing_file,
            test_read_args_simple,
            test_read_args_max_args_limit,
            test_read_args_max_cmdline_len_limit,
            test_transform_jar_path_elements,
            test_dedupe_hyphenated,
            test_truncate_jar,
            test_truncate_jar_short,
            test_is_unique_path_element,
            test_transform_multi_jars,
            test_tokenset,
            test_tokenset_overflow,
            test_extract_servicename_from_args_tomcat,
            test_extract_servicename_from_args_simple_jar,
            test_extract_servicename_from_args_module,
            test_extract_servicename_from_args_okhttp,
            test_extract_servicename_from_args_zk,
            test_extract_servicename_from_args_kafka,
            test_extract_servicename_from_args_spring,
            test_dots_to_dashes,
            test_env_var_already_set,
            test_is_legal_java_package_element,
            test_is_legal_java_fq_main_class,
            test_is_legal_java_module_main_class,
            test_is_legal_module,
            test_str_to_bool,
            test_disable_telemetry,
            test_enable_telemetry,
            test_enable_profiling,
            test_enable_profiling_memory,
            test_enable_metrics,
            test_concat_string_to_empty_just_enough_room,
            test_concat_string_to_empty_extra_room,
            test_concat_string_to_empty_not_enough_room,
            test_concat_string_to_nonempty_just_enough_room,
            test_long_cfg_attributes,
            test_auto_instrument_gen_svcname_disabled_but_specified
    };
    for (int i = 0; i < sizeof tests / sizeof tests[0]; ++i) {
        run_test(tests[i]);
    }
}

void run_test(test_func_t run_test) {
    unsetenv(java_tool_options_var);
    unsetenv(otel_service_name_var);
    unsetenv(resource_attributes_var);
    unsetenv(disable_env_var);
    logger l = new_logger();
    run_test(l);
    free_logger(l);
}

void test_auto_instrument_svc_name_specified(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    char *logs[256];
    int n = get_logs(l, logs);

    char *funcname = "test_auto_instrument_svc_name_specified";
    require_equal_ints(funcname, 4, n);
    require_equal_strings(funcname, "setting OTEL_SERVICE_NAME='my.override'", logs[0]);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[1]);
    require_equal_strings(funcname, "setting OTEL_RESOURCE_ATTRIBUTES='myattr=myval'", logs[2]);
    require_equal_strings(funcname, "sending metric", logs[3]);
    require_env(funcname, "-javaagent:/foo/asdf.jar", java_tool_options_var);
    require_env(funcname, "my.override", otel_service_name_var);
    cmdline_reader_close(cr);
}

void test_auto_instrument_gen_svc_name_explicitly_enabled(logger l) {
    char *funcname = "test_auto_instrument_gen_svc_name_explicitly_enabled";
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_generate_svcname_enabled, cr, fake_send_otlp_metric);
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_strings(funcname, "setting OTEL_SERVICE_NAME='foo'", logs[0]);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[1]);
    require_equal_strings(funcname, "setting OTEL_RESOURCE_ATTRIBUTES='myattr=myval'", logs[2]);
    require_equal_strings(funcname, "sending metric", logs[3]);
    require_equal_ints(funcname, 4, n);
    require_env(funcname, "foo", otel_service_name_var);
}

void test_auto_instrument_gen_svc_name_explicitly_disabled(logger l) {
    char *funcname = "test_auto_instrument_gen_svc_name_explicitly_disabled";
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_generate_svcname_disabled, cr, fake_send_otlp_metric);
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_strings(funcname, "service name generation explicitly disabled", logs[0]);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[1]);
    require_equal_strings(funcname, "sending metric", logs[2]);
    require_equal_ints(funcname, 3, n);
    require_unset_env(funcname, otel_service_name_var);
}

void test_auto_instrument_gen_svcname_disabled_but_specified(logger l) {
    char *funcname = "test_auto_instrument_gen_svcname_disabled_but_specified";
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_generate_svcname_disabled_but_explicitly_specified, cr, fake_send_otlp_metric);
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_strings(funcname, "service name generation explicitly disabled", logs[0]);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[1]);
    require_equal_strings(funcname, "sending metric", logs[2]);
    require_equal_ints(funcname, 3, n);
    require_unset_env(funcname, otel_service_name_var);
}

void test_auto_instrument_no_svc_name_in_config(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_no_svcname, cr, fake_send_otlp_metric);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_auto_instrument_no_svc_name_in_config";
    require_equal_ints(funcname, 3, n);
    require_equal_strings(funcname, "setting OTEL_SERVICE_NAME='foo'", logs[0]);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[1]);
    require_equal_strings(funcname, "sending metric", logs[2]);
    require_env(funcname, "-javaagent:/foo/asdf.jar", java_tool_options_var);
    require_env(funcname, "foo", otel_service_name_var);
    cmdline_reader_close(cr);
}

void test_auto_instrument_not_java(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "foo", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    char *funcname = "test_auto_instrument_not_java";
    require_unset_env(funcname, java_tool_options_var);
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_ints(funcname, 0, n);
    cmdline_reader_close(cr);
}

void test_auto_instrument_no_access(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_false, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    require_unset_env("test_auto_instrument_no_access", java_tool_options_var);
    char *logs[256];
    char *funcname = "test_auto_instrument_no_access";
    require_equal_ints(funcname, 1, get_logs(l, logs));
    require_equal_strings(funcname, "agent jar not found or no read access, quitting", logs[0]);
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_true(logger l) {
    setenv(disable_env_var, "true", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    require_unset_env("test_auto_instrument_splunk_env_var_true", "JAVA_TOOL_OPTIONS");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_false(logger l) {
    setenv(disable_env_var, "false", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    require_env("test_auto_instrument_splunk_env_var_false", "-javaagent:/foo/asdf.jar", "JAVA_TOOL_OPTIONS");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_false_caps(logger l) {
    setenv(disable_env_var, "FALSE", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    require_env("test_auto_instrument_splunk_env_var_false_caps", "-javaagent:/foo/asdf.jar", "JAVA_TOOL_OPTIONS");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_zero(logger l) {
    setenv(disable_env_var, "0", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr, fake_send_otlp_metric);
    require_env("test_auto_instrument_splunk_env_var_zero", "-javaagent:/foo/asdf.jar", "JAVA_TOOL_OPTIONS");
    cmdline_reader_close(cr);
}

void test_read_config_default(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, test_config_path);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config_default";
    require_equal_ints(funcname, 9, n);
    require_equal_strings(funcname, "reading config file: testdata/instrumentation-default.conf", logs[0]);
    require_equal_strings(funcname, "config: service_name not specified", logs[1]);
    require_equal_strings(funcname, "config: java_agent_jar=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", logs[2]);
    require_equal_strings(funcname, "config: resource_attributes=deployment.environment=test", logs[3]);
    require_equal_strings(funcname, "config: disable_telemetry not specified", logs[4]);
    require_equal_strings(funcname, "config: generate_service_name not specified", logs[5]);
    require_equal_strings(funcname, "config: enable_profiler not specified", logs[6]);
    require_equal_strings(funcname, "config: enable_profiler_memory not specified", logs[7]);
    require_equal_strings(funcname, "config: enable_metrics not specified", logs[8]);
    require_equal_strings(funcname, "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", cfg.java_agent_jar);
    require_equal_strings(funcname, NULL, cfg.service_name);
    require_equal_strings(funcname, "deployment.environment=test", cfg.resource_attributes);
    require_equal_strings(funcname, NULL, cfg.disable_telemetry);
    require_equal_strings(funcname, NULL, cfg.generate_service_name);
    require_equal_strings(funcname, NULL, cfg.enable_profiler);
    require_equal_strings(funcname, NULL, cfg.enable_profiler_memory);
    require_equal_strings(funcname, NULL, cfg.enable_metrics);
    free_config(&cfg);
}

void test_read_config_all_options(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, test_config_path_all_options);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config_all_options";
    require_equal_ints(funcname, 9, n);
    require_equal_strings(funcname, "reading config file: testdata/instrumentation-options.conf", logs[0]);
    require_equal_strings(funcname, "config: service_name=default.service", logs[1]);
    require_equal_strings(funcname, "config: java_agent_jar=/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", logs[2]);
    require_equal_strings(funcname, "config: resource_attributes=deployment.environment=test", logs[3]);
    require_equal_strings(funcname, "config: disable_telemetry=true", logs[4]);
    require_equal_strings(funcname, "config: generate_service_name=true", logs[5]);
    require_equal_strings(funcname, "config: enable_profiler=true", logs[6]);
    require_equal_strings(funcname, "config: enable_profiler_memory=true", logs[7]);
    require_equal_strings(funcname, "config: enable_metrics=true", logs[8]);

    require_equal_strings(funcname, "default.service", cfg.service_name);
    require_equal_strings(funcname, "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", cfg.java_agent_jar);
    require_equal_strings(funcname, "deployment.environment=test", cfg.resource_attributes);
    require_equal_strings(funcname, "true", cfg.disable_telemetry);
    require_equal_strings(funcname, "true", cfg.generate_service_name);
    require_equal_strings(funcname, "true", cfg.enable_profiler);
    require_equal_strings(funcname, "true", cfg.enable_profiler_memory);
    require_equal_strings(funcname, "true", cfg.enable_metrics);
    free_config(&cfg);
}

void test_read_config_missing_file(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, "foo.txt");
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config_missing_file";
    require_equal_ints(funcname, 9, n);
    require_equal_strings(funcname, "file not found: foo.txt", logs[0]);
    require_equal_strings(funcname, "config: service_name not specified", logs[1]);
    require_equal_strings(funcname, "config: java_agent_jar not specified", logs[2]);
    require_equal_strings(funcname, "config: resource_attributes not specified", logs[3]);
    require_equal_strings(funcname, "config: disable_telemetry not specified", logs[4]);
    require_equal_strings(funcname, "config: generate_service_name not specified", logs[5]);
    require_equal_strings(funcname, "config: enable_profiler not specified", logs[6]);
    require_equal_strings(funcname, "config: enable_profiler_memory not specified", logs[7]);
    require_equal_strings(funcname, "config: enable_metrics not specified", logs[8]);
    require_equal_strings(funcname, NULL, cfg.service_name);
    require_equal_strings(funcname, NULL, cfg.java_agent_jar);
    free_config(&cfg);
}

void test_read_args_simple(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    char *args[4];
    int n = get_cmdline_args(args, cr, 3, 128, NULL);
    char *funcname = "test_read_args_simple";
    require_equal_ints(funcname, 3, n);
    char *expected[] = {"java", "-jar", "foo.jar"};
    for (int i = 0; i < n; ++i) {
        require_equal_strings(funcname, expected[i], args[i]);
    }
    free_cmdline_args(args, n);
    cmdline_reader_close(cr);
}

void test_read_args_max_args_limit(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    char *args[1]; // not big enough!
    int n = get_cmdline_args(args, cr, 1, 128, NULL);
    char *funcname = "test_read_args_max_args_limit";
    require_equal_ints(funcname, 1, n);
    require_equal_strings(funcname, "java", args[0]);
    free_cmdline_args(args, n);
    cmdline_reader_close(cr);
}

void test_read_args_max_cmdline_len_limit(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    int max_args = 16;
    char *args[max_args];
    int max_cmdline_len = 8;
    int n = get_cmdline_args(args, cr, max_args, max_cmdline_len, l);
    char *funcname = "test_read_args_max_cmdline_len_limit";
    require_equal_ints(funcname, 1, n);
    require_equal_strings(funcname, "java", args[0]);
    free_cmdline_args(args, n);
    cmdline_reader_close(cr);
}

void test_is_unique_path_element(logger l) {
    require_true("test_is_unique_path_element", !is_unique_path_element("usr"));
    require_true("test_is_unique_path_element", !is_unique_path_element("bin"));
    require_true("test_is_unique_path_element", is_unique_path_element("foo"));
}

void test_truncate_jar(logger l) {
    char *jar_part = strdup("foo.jar");
    truncate_extension(jar_part);
    require_equal_strings("test_truncate_jar", "foo", jar_part);
    free(jar_part);
}

void test_truncate_jar_short(logger l) {
    char *jar_part = strdup(".jar");
    truncate_extension(jar_part);
    require_equal_strings("test_truncate_jar", ".jar", jar_part);
    free(jar_part);
}

void test_transform_jar_path_elements(logger l) {
    char path_buf[4096] = "";
    transform_jar_path_elements(path_buf, strdup("/usr/local/APACHE-TOMCAT/8.5.4/bin/apache-tomcat.jar"));
    char *funcname = "test_clean_jar_path_elements";
    require_equal_strings(funcname, "apache-tomcat-8.5.4-apache-tomcat", path_buf);
}

void test_dedupe_hyphenated(logger l) {
    char hyphen_buf[4096] = "";
    struct tokenset tks;
    init_tokenset(&tks);
    dedupe_hyphenated(hyphen_buf, strdup("apache-tomcat-8.5.4-apache-tomcat"), &tks);
    require_equal_strings("test_dedupe_hyphenated", "apache-tomcat-8.5.4", hyphen_buf);
    free_tokenset(&tks);
}

void test_transform_multi_jars(logger l) {
    char *arg = "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar:/usr/local/apache-tomcat/8.5.4/bin/tomcat-juli.jar";
    char transformed[4096] = "";
    struct tokenset tks;
    init_tokenset(&tks);
    transform_multi_jars(transformed, strdup(arg), &tks);
    require_equal_strings(
            "test_transform_multi_jars",
            "apache-tomcat-8.5.4-bootstrap-juli",
            transformed
    );
    free_tokenset(&tks);
}

void test_tokenset(logger l) {
    struct tokenset tks;
    init_tokenset(&tks);
    char *funcname = "test_tokenset";

    require_false(funcname, has_token(&tks, "foo"));
    add_token(&tks, "foo");
    require_true(funcname, has_token(&tks, "foo"));
    add_token(&tks, "foo");
    require_true(funcname, has_token(&tks, "foo"));

    require_false(funcname, has_token(&tks, "bar"));
    add_token(&tks, "bar");
    require_true(funcname, has_token(&tks, "bar"));
    require_false(funcname, has_token(&tks, "baz"));
    add_token(&tks, "baz");
    require_true(funcname, has_token(&tks, "baz"));
    add_token(&tks, "baz");
    require_true(funcname, has_token(&tks, "baz"));
    free_tokenset(&tks);
}

void test_tokenset_overflow(logger l) {
    struct tokenset tks;
    init_tokenset(&tks);
    char *funcname = "test_tokenset_overflow";
    for (int i = 1; i < TOKENSET_MAX_SIZE; ++i) {
        char buf[8];
        sprintf(buf, "tok-%d", i);
        require_false(funcname, has_token(&tks, buf));
        add_token(&tks, buf);
    }
    free_tokenset(&tks);
}

void test_extract_servicename_from_args_tomcat(logger l) {
    char *args[32];
    int num_args = tomcat_args(args);
    char service_name[4096] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_tomcat",
            "org-apache-catalina-startup-bootstrap",
            service_name
    );
}

void test_extract_servicename_from_args_simple_jar(logger l) {
    char *args[32];
    int num_args = petclinic_args(args);
    char service_name[4096] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_simple_jar",
            "app-spring-petclinic-2.4.5",
            service_name
    );
}

void test_extract_servicename_from_args_module(logger l) {
    char *args[32];
    int num_args = module_args(args);
    char service_name[4096] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings("test_extract_servicename_from_args_module", "com-mymodule-com-myorg-main", service_name);
}

void test_extract_servicename_from_args_okhttp(logger l) {
    char *args[32];
    int num_args = okhttp_and_jedis_args(args);
    char service_name[4096] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_module",
            "target-java-agent-example-1.0-snapshot-shaded",
            service_name
    );
}

void test_extract_servicename_from_args_zk(logger l) {
    char *args[32];
    int num_args = zk_args(args);
    char service_name[8192] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_zk",
            "org-apache-zookeeper-server-quorum-quorumpeermain",
            service_name
    );
}

void test_extract_servicename_from_args_kafka(logger l) {
    char *args[32];
    int num_args = kafka_args(args);
    char service_name[8192] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_kafka",
            "kafka-kafka",
            service_name
    );
}

void test_extract_servicename_from_args_spring(logger l) {
    char *args[32];
    int num_args = spring_args(args);
    char service_name[8192] = "";
    generate_servicename_from_args(service_name, args, num_args);
    require_equal_strings(
            "test_extract_servicename_from_args_spring",
            "com-example-demo-demoapplication",
            service_name
    );
}

void test_dots_to_dashes(logger l) {
    char *str = strdup("aaa.bbb.ccc");
    format_arg(str);
    require_equal_strings("test_dots_to_dashes", "aaa-bbb-ccc", str);
}

void test_env_var_already_set(logger l) {
    setenv("JAVA_TOOL_OPTIONS", "hello", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_svcname_explicitly_specified, cr,
                    fake_send_otlp_metric);
    char *funcname = "test_env_var_already_set";
    require_env(funcname, "hello", java_tool_options_var);
    cmdline_reader_close(cr);
}

void test_is_legal_java_package_element(logger l) {
    char *funcname = "test_is_legal_java_package_element";
    require_true(funcname, is_legal_java_package_element("zookeeper"));
    require_true(funcname, is_legal_java_package_element("has_underscores_42"));
    require_false(funcname, is_legal_java_package_element("has:colon"));
    require_false(funcname, is_legal_java_package_element("has/slash"));
    require_false(funcname, is_legal_java_package_element("has-dash"));
    require_false(funcname, is_legal_java_package_element("1_starts_with_number"));
}

void test_is_legal_java_module_main_class(logger l) {
    char *funcname = "is_legal_java_main_class_with_module";
    require_true(funcname, is_legal_java_main_class_with_module("com.my_module/com.myorg.Main"));
    require_true(funcname, is_legal_java_main_class_with_module("com.myorg.Main"));
    require_false(funcname, is_legal_java_main_class_with_module("http://my_module/com.myorg.Main"));
}

void test_is_legal_java_fq_main_class(logger l) {
    char *funcname = "test_is_legal_java_fq_main_class";
    require_true(funcname, is_legal_java_main_class("org.apache.zookeeper.server.quorum.QuorumPeerMain"));
    require_false(funcname, is_legal_java_main_class("org.apache.zookeeper.server.quorum.quorumPeerMain"));
    require_false(funcname, is_legal_java_main_class("org.apache.zookeeper-server.quorum.QuorumPeerMain"));
    require_false(funcname, is_legal_java_main_class("Main"));
    require_false(funcname, is_legal_java_main_class("http://google.com"));
}

void test_is_legal_module(logger l) {
    char *funcname = "test_is_legal_module";
    require_true(funcname, is_legal_module("foo.bar"));
    require_true(funcname, is_legal_module("bar"));
    require_false(funcname, is_legal_module("foo/bar"));
    require_false(funcname, is_legal_module("foo bar"));
}

void test_str_to_bool(logger l) {
    require_false("test_str_bool", str_to_bool("false", 0));
    require_false("test_str_bool", str_to_bool("FALSE", 0));
    require_false("test_str_bool", str_to_bool("0", 0));
    require_false("test_str_bool", str_to_bool(NULL, 0));
    require_true("test_str_bool", str_to_bool("true", 0));
    require_true("test_str_bool", str_to_bool("42", 0));
    require_true("test_str_bool", str_to_bool(NULL, 1));
}

void test_enable_telemetry(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_disable_telemetry_not_specified, cr,
                    fake_send_otlp_metric);
    char *logs[256];
    get_logs(l, logs);
    require_equal_strings("test_enable_telemetry", "sending metric", logs[2]);
}

void test_disable_telemetry(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_disable_telemetry_true, cr, fake_send_otlp_metric);
    char *logs[256];
    get_logs(l, logs);
    require_equal_strings("test_disable_telemetry", "disabling telemetry as per config", logs[2]);
}

void test_enable_profiling(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_enable_profiler, cr, fake_send_otlp_metric);
    require_env("test_enable_profiling", "-javaagent:/foo/asdf.jar -Dsplunk.profiler.enabled=true", "JAVA_TOOL_OPTIONS");
}

void test_enable_profiling_memory(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_enable_profiler_memory, cr, fake_send_otlp_metric);
    require_env("test_enable_profiling_memory", "-javaagent:/foo/asdf.jar -Dsplunk.profiler.memory.enabled=true", "JAVA_TOOL_OPTIONS");
}

void test_enable_metrics(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_enable_metrics, cr, fake_send_otlp_metric);
    require_env("test_enable_metrics", "-javaagent:/foo/asdf.jar -Dsplunk.metrics.enabled=true", "JAVA_TOOL_OPTIONS");
}

void test_concat_string_to_empty_just_enough_room(logger l) {
    char *funcname = "test_concat_string_to_empty_just_enough_room";
    char dest[4] = "";
    int res = concat_strings(dest, "abc", 4);
    require_equal_ints(funcname, 0, res);
    require_equal_strings(funcname, "abc", dest);
}

void test_concat_string_to_empty_extra_room(logger l) {
    char *funcname = "test_concat_string_to_empty_extra_room";
    char dest[8] = "";
    int res = concat_strings(dest, "abc", 8);
    require_equal_ints(funcname, 4, res);
    require_equal_strings(funcname, "abc", dest);
}

void test_concat_string_to_empty_not_enough_room(logger l) {
    char *funcname = "test_concat_string_to_empty_not_enough_room";
    char dest[2] = "";
    int res = concat_strings(dest, "abc", 2);
    require_equal_ints(funcname, -2, res);
    require_equal_strings(funcname, "a", dest);
}

void test_concat_string_to_nonempty_just_enough_room(logger l) {
    char *funcname = "test_concat_string_to_nonempty_just_enough_room";
    char dest[4] = "ab";
    int res = concat_strings(dest, "c", 4);
    require_equal_ints(funcname, 0, res);
    require_equal_strings(funcname, "abc", dest);
}

void test_long_cfg_attributes(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_config_max_attributes, cr, fake_send_otlp_metric);
    require_env_len("test_long_cfg_attributes", 255, "OTEL_SERVICE_NAME");
    require_env_len("test_long_cfg_attributes", 365, "JAVA_TOOL_OPTIONS");
    require_env_len("test_long_cfg_attributes", 255, "OTEL_RESOURCE_ATTRIBUTES");
}

// fakes/testdata

void fake_send_otlp_metric(logger log, char *service_name) {
    log_debug(log, "sending metric");
}

void fake_config_svcname_explicitly_specified(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->service_name = strdup("my.override");
    cfg->resource_attributes = strdup("myattr=myval");
    cfg->disable_telemetry = strdup("false");
}

void fake_config_generate_svcname_enabled(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->resource_attributes = strdup("myattr=myval");
    cfg->disable_telemetry = strdup("false");
    cfg->generate_service_name = strdup("true");
}

void fake_config_generate_svcname_disabled(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->generate_service_name = strdup("false");
}

void fake_config_generate_svcname_disabled_but_explicitly_specified(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->generate_service_name = strdup("false");
    cfg->service_name = strdup("my-explicit-servicename");
}

void fake_config_no_svcname(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
}

void fake_config_disable_telemetry_not_specified(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
}

void fake_config_disable_telemetry_true(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->disable_telemetry = strdup("true");
}

void fake_config_enable_profiler(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->enable_profiler = strdup("true");
}

void fake_config_enable_profiler_memory(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->enable_profiler_memory = strdup("true");
}

void fake_config_enable_metrics(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->enable_metrics = strdup("true");
}

void fake_config_max_attributes(logger log, struct config *cfg, char *path) {
    char *str_255 = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
                    "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
                    "1234567890123456789012345678901234567890123456789012345";
    cfg->java_agent_jar = strdup(str_255);
    cfg->service_name = strdup(str_255);
    cfg->resource_attributes = strdup(str_255);
    cfg->enable_profiler = strdup("true");
    cfg->enable_profiler_memory = strdup("true");
    cfg->enable_metrics = strdup("true");
}

cmdline_reader new_default_test_cmdline_reader() {
    char cmdline[] = {'j', 'a', 'v', 'a', 0, '-', 'j', 'a', 'r', 0, 'f', 'o', 'o', '.', 'j', 'a', 'r', 0};
    return new_test_cmdline_reader(cmdline, sizeof(cmdline));
}

int access_check_true(const char *s) {
    return 1;
}

int access_check_false(const char *s) {
    return 0;
}

int tomcat_args(char *args[]) {
    int n = 0;
    args[n++] = "java";
    args[n++] = "-Djava.util.logging.config.file=/usr/local/apache-tomcat/8.5.4/conf/logging.properties";
    args[n++] = "-Djava.util.logging.manager=org.apache.juli.ClassLoaderLogManager";
    args[n++] = "-Djdk.tls.ephemeralDHKeySize=2048";
    args[n++] = "-classpath";
    args[n++] = "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar:/usr/local/apache-tomcat/8.5.4/bin/tomcat-juli.jar";
    args[n++] = "-Dcatalina.base=/usr/local/apache-tomcat/8.5.4";
    args[n++] = "-Dcatalina.home=/usr/local/apache-tomcat/8.5.4";
    args[n++] = "-Djava.io.tmpdir=/usr/local/apache-tomcat/8.5.4/temp";
    args[n++] = "org.apache.catalina.startup.Bootstrap";
    args[n++] = "start";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

int petclinic_args(char *args[]) {
    int n = 0;
    args[n++] = "/usr/bin/java";
    args[n++] = "-cp";
    args[n++] = "my/classpath";
    args[n++] = "-jar";
    args[n++] = "app/spring-petclinic-2.4.5.jar";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

int module_args(char *args[]) {
    int n = 0;
    args[n++] = "/usr/bin/java";
    args[n++] = "--module-path";
    args[n++] = "mods";
    args[n++] = "-m";
    args[n++] = "com.mymodule/com.myorg.Main";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

int okhttp_and_jedis_args(char *args[]) {
    int n = 0;
    args[n++] = "java";
    args[n++] = "-jar";
    args[n++] = "target/java-agent-example-1.0-SNAPSHOT-shaded.jar";
    args[n++] = "https://google.com";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

int zk_args(char *args[]) {
    int n = 0;
    args[n++] = "java";
    args[n++] = "-Xmx512M";
    args[n++] = "-Xms512M";
    args[n++] = "-server";
    args[n++] = "-XX:+UseG1GC";
    args[n++] = "-XX:MaxGCPauseMillis=20";
    args[n++] = "-XX:InitiatingHeapOccupancyPercent=35";
    args[n++] = "-XX:+ExplicitGCInvokesConcurrent";
    args[n++] = "-XX:MaxInlineLevel=15";
    args[n++] = "-Djava.awt.headless=true";
    args[n++] = "-Xlog:gc*:file=/usr/local/kafka/kafka_2.13-3.1.0/bin/../logs/zookeeper-gc.log:time,tags:filecount=10,filesize=100M";
    args[n++] = "-Dcom.sun.management.jmxremote";
    args[n++] = "-Dcom.sun.management.jmxremote.authenticate=false";
    args[n++] = "-Dcom.sun.management.jmxremote.ssl=false";
    args[n++] = "-Dkafka.logs.dir=/usr/local/kafka/kafka_2.13-3.1.0/bin/../logs";
    args[n++] = "-Dlog4j.configuration=file:bin/../config/log4j.properties";
    args[n++] = "-cp";
    args[n++] = zk_classpath();
    args[n++] = "org.apache.zookeeper.server.quorum.QuorumPeerMain";
    args[n++] = "config/zookeeper.properties";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

char *zk_classpath() {
    return "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/activation-1.1.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/aopalliance-repackaged-2.6.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/argparse4j-0.7.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/audience-annotations-0.5.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/commons-cli-1.4.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/commons-lang3-3.8.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-api-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-basic-auth-extension-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-file-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-json-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-mirror-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-mirror-client-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-runtime-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-transforms-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-api-2.6.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-locator-2.6.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-utils-2.6.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-annotations-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-core-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-databind-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-dataformat-csv-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-datatype-jdk8-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-jaxrs-base-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-jaxrs-json-provider-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-module-jaxb-annotations-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-module-scala_2.13-2.12.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.activation-api-1.2.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.annotation-api-1.3.5.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.inject-2.6.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.validation-api-2.0.2.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.ws.rs-api-2.1.6.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.xml.bind-api-2.3.2.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javassist-3.27.0-GA.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javax.servlet-api-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javax.ws.rs-api-2.1.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jaxb-api-2.3.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-client-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-common-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-container-servlet-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-container-servlet-core-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-hk2-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-server-2.34.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-client-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-continuation-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-http-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-io-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-security-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-server-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-servlet-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-servlets-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-util-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-util-ajax-9.4.43.v20210629.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jline-3.12.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jopt-simple-5.0.4.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jose4j-0.7.8.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka_2.13-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-clients-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-log4j-appender-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-metadata-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-raft-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-server-common-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-shell-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-storage-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-storage-api-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-examples-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-scala_2.13-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-test-utils-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-tools-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/log4j-1.2.17.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/lz4-java-1.8.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/maven-artifact-3.8.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/metrics-core-2.2.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/metrics-core-4.1.12.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-buffer-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-codec-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-common-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-handler-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-resolver-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-native-epoll-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-native-unix-common-4.1.68.Final.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/osgi-resource-locator-1.0.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/paranamer-2.8.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/plexus-utils-3.2.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/reflections-0.9.12.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/rocksdbjni-6.22.1.1.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-collection-compat_2.13-2.4.4.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-java8-compat_2.13-1.0.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-library-2.13.6.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-logging_2.13-3.9.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-reflect-2.13.6.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/slf4j-api-1.7.30.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/slf4j-log4j12-1.7.30.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/snappy-java-1.1.8.4.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/trogdor-3.1.0.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zookeeper-3.6.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zookeeper-jute-3.6.3.jar:"
           "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zstd-jni-1.5.0-4.jar";
}

int kafka_args(char *args[]) {
    int n = 0;
    args[n++] = "java";
    args[n++] = "-Xmx1G";
    args[n++] = "-Xms1G";
    args[n++] = "-server";
    args[n++] = "-XX:+UseG1GC";
    args[n++] = "-XX:MaxGCPauseMillis=20";
    args[n++] = "-XX:InitiatingHeapOccupancyPercent=35";
    args[n++] = "-XX:+ExplicitGCInvokesConcurrent";
    args[n++] = "-XX:MaxInlineLevel=15";
    args[n++] = "-Djava.awt.headless=true";
    args[n++] = "-Xlog:gc*:file=/usr/local/kafka/kafka_2.13-3.1.0/bin/../logs/kafkaServer-gc.log:time,tags:filecount=10,filesize=100M";
    args[n++] = "-Dcom.sun.management.jmxremote";
    args[n++] = "-Dcom.sun.management.jmxremote.authenticate=false";
    args[n++] = "-Dcom.sun.management.jmxremote.ssl=false";
    args[n++] = "-Dkafka.logs.dir=/usr/local/kafka/kafka_2.13-3.1.0/bin/../logs";
    args[n++] = "-Dlog4j.configuration=file:bin/../config/log4j.properties";
    args[n++] = "-cp";
    args[n++] = "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/activation-1.1.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/aopalliance-repackaged-2.6.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/argparse4j-0.7.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/audience-annotations-0.5.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/commons-cli-1.4.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/commons-lang3-3.8.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-api-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-basic-auth-extension-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-file-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-json-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-mirror-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-mirror-client-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-runtime-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/connect-transforms-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-api-2.6.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-locator-2.6.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/hk2-utils-2.6.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-annotations-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-core-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-databind-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-dataformat-csv-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-datatype-jdk8-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-jaxrs-base-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-jaxrs-json-provider-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-module-jaxb-annotations-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jackson-module-scala_2.13-2.12.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.activation-api-1.2.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.annotation-api-1.3.5.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.inject-2.6.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.validation-api-2.0.2.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.ws.rs-api-2.1.6.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jakarta.xml.bind-api-2.3.2.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javassist-3.27.0-GA.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javax.servlet-api-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/javax.ws.rs-api-2.1.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jaxb-api-2.3.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-client-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-common-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-container-servlet-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-container-servlet-core-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-hk2-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jersey-server-2.34.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-client-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-continuation-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-http-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-io-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-security-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-server-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-servlet-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-servlets-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-util-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jetty-util-ajax-9.4.43.v20210629.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jline-3.12.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jopt-simple-5.0.4.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/jose4j-0.7.8.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka_2.13-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-clients-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-log4j-appender-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-metadata-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-raft-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-server-common-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-shell-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-storage-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-storage-api-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-examples-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-scala_2.13-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-streams-test-utils-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/kafka-tools-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/log4j-1.2.17.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/lz4-java-1.8.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/maven-artifact-3.8.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/metrics-core-2.2.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/metrics-core-4.1.12.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-buffer-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-codec-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-common-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-handler-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-resolver-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-native-epoll-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/netty-transport-native-unix-common-4.1.68.Final.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/osgi-resource-locator-1.0.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/paranamer-2.8.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/plexus-utils-3.2.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/reflections-0.9.12.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/rocksdbjni-6.22.1.1.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-collection-compat_2.13-2.4.4.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-java8-compat_2.13-1.0.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-library-2.13.6.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-logging_2.13-3.9.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/scala-reflect-2.13.6.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/slf4j-api-1.7.30.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/slf4j-log4j12-1.7.30.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/snappy-java-1.1.8.4.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/trogdor-3.1.0.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zookeeper-3.6.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zookeeper-jute-3.6.3.jar:"
                "/usr/local/kafka/kafka_2.13-3.1.0/bin/../libs/zstd-jni-1.5.0-4.jar";
    args[n++] = "kafka.Kafka";
    args[n++] = "config/server.properties";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

int spring_args(char *args[]) {
    int n = 0;
    args[n++] = "java";
    args[n++] = "-cp";
    args[n++] = "/home/pcollins/spring/demo/target/classes:"
                "/usr/lib/m2/repository/org/springframework/boot/spring-boot/2.6.7/spring-boot-2.6.7.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-context/5.3.19/spring-context-5.3.19.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-aop/5.3.19/spring-aop-5.3.19.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-beans/5.3.19/spring-beans-5.3.19.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-expression/5.3.19/spring-expression-5.3.19.jar:"
                "/usr/lib/m2/repository/org/springframework/boot/spring-boot-autoconfigure/2.6.7/spring-boot-autoconfigure-2.6.7.jar:"
                "/usr/lib/m2/repository/ch/qos/logback/logback-classic/1.2.11/logback-classic-1.2.11.jar:"
                "/usr/lib/m2/repository/ch/qos/logback/logback-core/1.2.11/logback-core-1.2.11.jar:"
                "/usr/lib/m2/repository/org/apache/logging/log4j/log4j-to-slf4j/2.17.2/log4j-to-slf4j-2.17.2.jar:"
                "/usr/lib/m2/repository/org/apache/logging/log4j/log4j-api/2.17.2/log4j-api-2.17.2.jar:"
                "/usr/lib/m2/repository/org/slf4j/jul-to-slf4j/1.7.36/jul-to-slf4j-1.7.36.jar:"
                "/usr/lib/m2/repository/jakarta/annotation/jakarta.annotation-api/1.3.5/jakarta.annotation-api-1.3.5.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-core/5.3.19/spring-core-5.3.19.jar:"
                "/usr/lib/m2/repository/org/springframework/spring-jcl/5.3.19/spring-jcl-5.3.19.jar:"
                "/usr/lib/m2/repository/org/yaml/snakeyaml/1.29/snakeyaml-1.29.jar:"
                "/usr/lib/m2/repository/org/slf4j/slf4j-api/1.7.36/slf4j-api-1.7.36.jar";
    args[n++] = "com.example.demo.DemoApplication";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}

