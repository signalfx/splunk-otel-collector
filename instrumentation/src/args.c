#include "args.h"
#include "cmdline_reader.h"
#include <string.h>
#include <stdlib.h>
#include <ctype.h>

// 4096 is the max command line statement length on linux
static const int max_statement_len = 4096;

// individual args on the heap and must be freed
int get_cmdline_args(char **args, int max_args, cmdline_reader cr) {
    int args_idx = 0;
    int arg_char_offset = 0;
    char arg[max_statement_len];
    cmdline_reader_open(cr);
    while (!cmdline_reader_is_eof(cr)) {
        char c = cmdline_reader_get_char(cr);
        arg[arg_char_offset] = c;
        arg_char_offset += 1;
        if (c == 0) {
            args[args_idx] = strdup(arg);
            args_idx += 1;
            if (args_idx == max_args) {
                break;
            }
            arg_char_offset = 0;
        }
    }
    return args_idx;
}

void free_cmdline_args(char **args, int num_args) {
    for (int i = 0; i < num_args; ++i) {
        free(args[i]);
    }
}

void format_arg(char *str) {
    for (; *str != 0; ++str) {
        if (*str == '.' || *str == '/') {
            *str = '-';
        } else {
            *str = (char) tolower(*str);
        }
    }
}

// individual `arg` strings are on the heap
void generate_servicename_from_args(char *service_name, char **args, int num_args) {
    for (int i = 0; i < num_args; ++i) {
        char *arg = args[i];
        if (strstr(arg, ".jar") != NULL) {
            concat_jars_arg(service_name, arg);
        }
        if (arg[0] == '-') {
            continue;
        }
        if (strstr(arg, ".") != NULL) {
            format_arg(arg);
            strcpy(service_name, arg);
        }
    }
}

// `concat_jars_arg` is on the stack, `arg` is on the heap
// `arg` is e.g. "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar:/usr/local/apache-tomcat/8.5.4/bin/tomcat-juli.jar"
void concat_jars_arg(char *cleaned_jars_arg, char *arg) {
    char *token;
    while ((token = strsep(&arg, ":")) != NULL) {
        clean_jar_path(cleaned_jars_arg, token);
    }
}

void clean_jar_path(char *cleaned_str, char *path) {
    // path = e.g. "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar"
    char *token;
    while ((token = strsep(&path, "/")) != NULL) {
        if (is_unique_path_element(token)) {
            if (strlen(cleaned_str) > 0) {
                strcat(cleaned_str, "-");
            }
            truncate_extension(token);
            strcat(cleaned_str, token);
        }
    }
}

bool is_unique_path_element(char *path_element) {
    if (strlen(path_element) == 0) {
        return false;
    }

    static const char *standard_path_parts[] = {"usr", "local", "bin", "home", "etc", "lib", "opt"};
    static const int num_standard_path_parts = 7;
    for (int i = 0; i < num_standard_path_parts; ++i) {
        const char *part = standard_path_parts[i];
        if (strcmp(part, path_element) == 0) {
            return false;
        }
    }
    return true;
}

// removes a .jar suffix/extension from a string if it's long enough
void truncate_extension(char *str) {
    unsigned long len = strlen(str);
    if (len <= 4) {
        return;
    }
    if (strstr(str + len - 4, ".jar")) {
        str[len - 4] = 0;
    }
}
