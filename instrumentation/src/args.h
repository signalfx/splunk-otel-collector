#ifndef INSTRUMENTATION_ARGS_H
#define INSTRUMENTATION_ARGS_H

#include "cmdline_reader.h"
#include "logger.h"

#define TOKENSET_MAX_SIZE 256

struct tokenset {
    int i;
    char *tokens[TOKENSET_MAX_SIZE];
};

void init_tokenset(struct tokenset *tks);

void free_tokenset(struct tokenset *tks);

int has_token(struct tokenset *tks, char *token);

void add_token(struct tokenset *tks, char *token);

int get_cmdline_args(char **args, cmdline_reader cr, int max_args, int max_cmdline_len, logger log);

void free_cmdline_args(char **args, int num_args);

void generate_servicename_from_args(char *dest, char **args, int num_args);

int is_legal_java_main_class(const char *str);

int is_capital_letter(char ch);

int is_legal_java_main_class_with_module(const char *str);

int is_legal_module(char *module);

int is_legal_java_package_element(const char *str);

void transform_multi_jars(char *dest, char *arg, struct tokenset *tks);

void transform_jar_path_elements(char *out, char *path);

void dedupe_hyphenated(char *out, char *str, struct tokenset *pTokenset);

int is_unique_path_element(char *path_element);

void truncate_extension(char *str);

void format_arg(char *str);

#endif //INSTRUMENTATION_ARGS_H
