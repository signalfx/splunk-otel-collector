#include "args.h"
#include "cmdline_reader.h"
#include <string.h>
#include <stdlib.h>
#include <ctype.h>
#include <stdio.h>

#define MAX_ARG_LEN 4096

// individual args are copied onto the heap and should be freed
int get_cmdline_args(char **args, cmdline_reader cr, int max_args, int max_cmdline_len, logger log) {
    int cmdline_idx = 0;
    int args_idx = 0;
    int arg_char_offset = 0;
    char arg[MAX_ARG_LEN];
    cmdline_reader_open(cr);
    while (!cmdline_reader_is_eof(cr)) {
        if (cmdline_idx++ == max_cmdline_len) {
            log_warning(log, "command line too long, truncating");
            break;
        }
        char c = cmdline_reader_get_char(cr);
        arg[arg_char_offset] = c;
        arg_char_offset += 1;
        if (c == 0 || arg_char_offset == MAX_ARG_LEN) {
            args[args_idx] = strndup(arg, MAX_ARG_LEN);
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

void init_tokenset(struct tokenset *tks) {
    tks->i = 0;
}

void free_tokenset(struct tokenset *tks) {
    for (int i = 0; i < tks->i; ++i) {
        free(tks->tokens[i]);
    }
}

void add_token(struct tokenset *tks, char *token) {
    // it is unlikely, but if we have too many tokens, just stop adding them
    if (tks->i < TOKENSET_MAX_SIZE) {
        tks->tokens[tks->i++] = strdup(token);
    }
}

bool has_token(struct tokenset *tks, char *token) {
    // not doing a set implementation at this time since size of array is small
    for (int i = 0; i < tks->i; ++i) {
        if (strcmp(tks->tokens[i], token) == 0) {
            return true;
        }
    }
    return false;
}

void generate_servicename_from_args(char *out, char **args, int num_args) {
    struct tokenset tks;
    init_tokenset(&tks);
    for (int i = 0; i < num_args; ++i) {
        char *arg = args[i];
        if (strstr(arg, ".jar") != NULL) {
            transform_multi_jars(out, arg, &tks);
        }
        if (arg[0] == '-') {
            continue;
        }
        if (is_legal_java_main_class_with_module(arg)) {
            format_arg(arg);
            strcpy(out, arg);
            return;
        }
    }
    free_tokenset(&tks);
}

// concatenates colon separated jars and removes non-uniquely-identifying dirs as well as double dots from the path
// `arg` is e.g. "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar:/usr/local/apache-tomcat/8.5.4/bin/tomcat-juli.jar"
// `out` is on the stack, `arg` is on the heap
void transform_multi_jars(char *out, char *arg, struct tokenset *tks) {
    char *token;
    while ((token = strsep(&arg, ":")) != NULL) {
        char transformed_jar_path_elements[MAX_ARG_LEN] = "";
        transform_jar_path_elements(transformed_jar_path_elements, token);

        char deduped[MAX_ARG_LEN] = "";
        dedupe_hyphenated(deduped, transformed_jar_path_elements, tks);

        if (strlen(out) > 0) {
            strcat(out, "-");
        }
        strcat(out, deduped);
    }
}

void tolowerstr(char *str);

void transform_jar_path_elements(char *out, char *path) {
    // path = e.g. "/usr/local/apache-tomcat/8.5.4/bin/bootstrap.jar"
    char *token;
    while ((token = strsep(&path, "/")) != NULL) {
        if (!is_unique_path_element(token)) {
            continue;
        }
        if (strlen(out) > 0) {
            strcat(out, "-");
        }
        truncate_extension(token);
        tolowerstr(token);
        strcat(out, token);
    }
}

void tolowerstr(char *str) {
    for (int i = 0; i < strlen(str); ++i) {
        str[i] = (char) tolower(str[i]);
    }
}

void dedupe_hyphenated(char *out, char *str, struct tokenset *pTokenset) {
    char *tok;
    while ((tok = strsep(&str, "-")) != NULL) {
        if (has_token(pTokenset, tok)) {
            continue;
        }
        add_token(pTokenset, tok);
        if (strlen(out) > 0) {
            strcat(out, "-");
        }
        strcat(out, tok);
    }
}

bool is_unique_path_element(char *path_element) {
    if (strlen(path_element) == 0) {
        return false;
    }

    static const char *standard_path_parts[] = {"usr", "local", "bin", "home", "etc", "lib", "opt", ".."};
    static const int num_standard_path_parts = 8;
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

bool is_legal_java_main_class_with_module(const char *str) {
    int num_slashes = 0;
    int num_dots = 0;
    for (int i = 0; str[i] != 0; ++i) {
        if (str[i] == '.') {
            ++num_dots;
        }
        if (str[i] == '/') {
            ++num_slashes;
        }
    }
    if (num_dots == 0) {
        return false;
    }
    if (num_slashes == 0) {
        return is_legal_java_main_class(str);
    } else if (num_slashes > 1) {
        return false;
    }
    char *fq_main_package = strdup(str);
    char *module = strsep(&fq_main_package, "/");

    if (!is_legal_java_main_class(fq_main_package)) {
        return false;
    }

    if (!is_legal_module(module)) {
        printf("not legal! module: %s\n", module);
        return false;
    }

    return true;
}

bool is_legal_java_main_class(const char *str) {
    if (strstr(str, ".") == NULL) {
        return false;
    }
    char *dup = strdup(str);
    char *prev;
    while (true) {
        char *token = strsep(&dup, ".");
        if (token == NULL) {
            return is_capital_letter(prev[0]);
        }
        if (!is_legal_java_package_element(token)) {
            return false;
        }
        prev = token;
    }
}

bool is_capital_letter(const char ch) {
    return ch >= 'A' && ch <= 'Z';
}

bool is_legal_module(char *module) {
    char *dup = strdup(module);
    char *token;
    while ((token = strsep(&dup, ".")) != NULL) {
        if (!is_legal_java_package_element(token)) {
            return false;
        }
    }
    return true;
}

// tests if the parts between the dots in e.g. some.package.MyMain are legal
bool is_legal_java_package_element(const char *str) {
    for (int i = 0;; ++i) {
        char ch = str[i];
        if (ch == 0) {
            break;
        }
        if (i == 0 && ch >= '0' && ch <= '9') {
            return false;
        }
        if (ch < '0' || (ch > '9' && ch < 'A') || (ch > 'Z' && ch < '_') || ch == '`' || ch > 'z') {
            return false;
        }
    }
    return true;
}
