#ifndef INSTRUMENTATION_ARGS_H
#define INSTRUMENTATION_ARGS_H

#include "cmdline_reader.h"

int get_cmdline_args(char **args, int max_args, cmdline_reader cr);

void free_cmdline_args(char **args, int num_args);

void generate_servicename_from_args(char *service_name, char **args, int num_args);

void concat_jars_arg(char *cleaned_jars_arg, char *arg);

void clean_jar_path(char *cleaned_str, char *path);

bool is_unique_path_element(char *path_element);

void truncate_extension(char *str);

void format_arg(char *str);

#endif //INSTRUMENTATION_ARGS_H
