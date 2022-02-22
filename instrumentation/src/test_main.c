#include "test_main.h"
#include "splunk.h"
#include "config.h"
#include "args.h"
#include "test_utils.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static char *const test_config_path = "testdata/instrumentation.conf";

int main(void) {
    run_tests();
    puts("PASS");
    return EXIT_SUCCESS;
}

void run_tests() {
    test_func_t *tests[] = {
            test_auto_instrument_svc_name_in_config,
            test_auto_instrument_no_svc_name_in_config,
            test_auto_instrument_not_java,
            test_auto_instrument_no_access,
            test_auto_instrument_splunk_env_var_true,
            test_auto_instrument_splunk_env_var_false,
            test_auto_instrument_splunk_env_var_false_caps,
            test_auto_instrument_splunk_env_var_zero,
            test_read_config,
            test_read_config_missing_file,
            test_read_args_simple,
            test_read_args_limit,
            test_clean_jar,
            test_truncate_jar,
            test_truncate_jar_short,
            test_is_unique_path_element,
            test_clean_jars,
            test_extract_servicename_from_args_tomcat,
            test_extract_servicename_from_args_simple_jar,
            test_extract_servicename_from_args_module,
            test_dots_to_dashes,
            test_env_var_already_set
    };
    for (int i = 0; i < sizeof tests / sizeof tests[0]; ++i) {
        run_test(tests[i]);
    }
}

void run_test(test_func_t run_test) {
    unsetenv("JAVA_TOOL_OPTIONS");
    unsetenv("OTEL_SERVICE_NAME");
    unsetenv("DISABLE_SPLUNK_AUTOINSTRUMENTATION");
    logger l = new_logger();
    run_test(l);
    free_logger(l);
}

void test_auto_instrument_svc_name_in_config(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_auto_instrument_svc_name_in_config";
    require_equal_ints(funcname, 2, n);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[0]);
    require_equal_strings(funcname, "setting OTEL_SERVICE_NAME='my.service'", logs[1]);
    require_equal_strings(funcname, "-javaagent:/foo/asdf.jar", getenv("JAVA_TOOL_OPTIONS"));
    require_equal_strings(funcname, "my.service", getenv("OTEL_SERVICE_NAME"));
    cmdline_reader_close(cr);
}

void test_auto_instrument_no_svc_name_in_config(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config_no_svcname, cr);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_auto_instrument_no_svc_name_in_config";
    require_equal_ints(funcname, 2, n);
    require_equal_strings(funcname, "setting JAVA_TOOL_OPTIONS='-javaagent:/foo/asdf.jar'", logs[0]);
    require_equal_strings(funcname, "setting OTEL_SERVICE_NAME='foo'", logs[1]);
    require_equal_strings(funcname, "-javaagent:/foo/asdf.jar", getenv("JAVA_TOOL_OPTIONS"));
    require_equal_strings(funcname, "foo", getenv("OTEL_SERVICE_NAME"));
    cmdline_reader_close(cr);
}

void test_auto_instrument_not_java(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "foo", fake_load_config, cr);
    char *env = getenv(java_tool_options_var);
    if (env) {
        fail();
    }
    char *logs[256];
    int n = get_logs(l, logs);
    require_equal_ints("test_auto_instrument_not_java", 0, n);
    cmdline_reader_close(cr);
}

void test_auto_instrument_no_access(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_false, "java", fake_load_config, cr);
    char *env = getenv(java_tool_options_var);
    if (env) {
        fail();
    }
    char *logs[256];
    char *funcname = "test_auto_instrument_no_access";
    require_equal_ints(funcname, 1, get_logs(l, logs));
    require_equal_strings(funcname, "agent jar not found or no read access, quitting", logs[0]);
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_true(logger l) {
    setenv(disable_env_var_name, "true", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    require_unset_env("test_auto_instrument_splunk_env_var_true", "JAVA_TOOL_OPTIONS");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_false(logger l) {
    setenv(disable_env_var_name, "false", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    require_env("test_auto_instrument_splunk_env_var_false", "JAVA_TOOL_OPTIONS", "-javaagent:/foo/asdf.jar");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_false_caps(logger l) {
    setenv(disable_env_var_name, "FALSE", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    require_env("test_auto_instrument_splunk_env_var_false_caps", "JAVA_TOOL_OPTIONS", "-javaagent:/foo/asdf.jar");
    cmdline_reader_close(cr);
}

void test_auto_instrument_splunk_env_var_zero(logger l) {
    setenv(disable_env_var_name, "0", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    require_env("test_auto_instrument_splunk_env_var_zero", "JAVA_TOOL_OPTIONS", "-javaagent:/foo/asdf.jar");
    cmdline_reader_close(cr);
}

void test_read_config(logger l) {
    struct config cfg = {.java_agent_jar = NULL, .service_name = NULL};
    load_config(l, &cfg, test_config_path);
    char *logs[256];
    int n = get_logs(l, logs);
    char *funcname = "test_read_config";
    require_equal_ints(funcname, 1, n);
    require_equal_strings(funcname, "reading config file: testdata/instrumentation.conf", logs[0]);
    require_equal_strings(funcname, "default.service", cfg.service_name);
    require_equal_strings(funcname, "/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar", cfg.java_agent_jar);
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
    require_equal_strings(funcname, "service_name not found in config", logs[1]);
    require_equal_strings(funcname, "java_agent_jar not found in config", logs[2]);
    require_equal_strings(funcname, NULL, cfg.service_name);
    require_equal_strings(funcname, NULL, cfg.java_agent_jar);
    free_config(&cfg);
}

void test_read_args_simple(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    char *args[4];
    int n = get_cmdline_args(args, 3, cr);
    char *funcname = "test_read_args_simple";
    require_equal_ints(funcname, 3, n);
    char *expected[] = {"java", "-jar", "foo.jar"};
    for (int i = 0; i < n; ++i) {
        require_equal_strings(funcname, expected[i], args[i]);
    }
    free_cmdline_args(args, n);
    cmdline_reader_close(cr);
}

void test_read_args_limit(logger l) {
    cmdline_reader cr = new_default_test_cmdline_reader();
    char *args[1]; // not big enough!
    int n = get_cmdline_args(args, 1, cr);
    char *funcname = "test_read_args_limit";
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

void test_clean_jar(logger l) {
    char cleaned[4096] = "";
    clean_jar_path(cleaned, strdup("/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar"));
    require_equal_strings("test_clean_jar", "apache-tomcat-8.5.4-bootstrap", cleaned);
}

void test_clean_jars(logger l) {
    char *arg = "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar:/usr/local/apache-tomcat/8.5.4/bin/tomcat-juli.jar";
    char cleaned[4096] = "";
    concat_jars_arg(cleaned, strdup(arg));
    require_equal_strings(
            "test_clean_jars",
            "apache-tomcat-8.5.4-bootstrap-apache-tomcat-8.5.4-tomcat-juli",
            cleaned
    );
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
    require_equal_strings("test_extract_servicename_from_args_module", "com-my-module-com-myorg-main", service_name);
}

void test_dots_to_dashes(logger l) {
    char *str = strdup("aaa.bbb.ccc");
    format_arg(str);
    require_equal_strings("test_dots_to_dashes", "aaa-bbb-ccc", str);
}

void test_env_var_already_set(logger l) {
    setenv("JAVA_TOOL_OPTIONS", "asdf", 0);
    cmdline_reader cr = new_default_test_cmdline_reader();
    auto_instrument(l, access_check_true, "java", fake_load_config, cr);
    require_env("test_env_var_already_set", "JAVA_TOOL_OPTIONS", "asdf");
    cmdline_reader_close(cr);
}

// fakes/testdata

void fake_load_config(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
    cfg->service_name = strdup("my.service");
}

void fake_load_config_no_svcname(logger log, struct config *cfg, char *path) {
    cfg->java_agent_jar = strdup("/foo/asdf.jar");
}

cmdline_reader new_default_test_cmdline_reader() {
    char cmdline[] = {'j', 'a', 'v', 'a', 0, '-', 'j', 'a', 'r', 0, 'f', 'o', 'o', '.', 'j', 'a', 'r', 0};
    return new_test_cmdline_reader(cmdline, sizeof(cmdline));
}

bool access_check_true(const char *s) {
    return true;
}

bool access_check_false(const char *s) {
    return false;
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
    args[n++] = "-jar";
    args[n++] = "-p";
    args[n++] = "my/module/path";
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
    args[n++] = "com.my-module/com.myorg.Main";
    for (int i = 0; i < n; ++i) {
        args[i] = strdup(args[i]);
    }
    return n;
}
