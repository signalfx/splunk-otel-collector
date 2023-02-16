#ifndef INSTRUMENTATION_CMDLINE_READER_H
#define INSTRUMENTATION_CMDLINE_READER_H

typedef struct cmdline_reader_impl *cmdline_reader;
cmdline_reader new_cmdline_reader();
cmdline_reader new_test_cmdline_reader(char *cmdline, int len);
void cmdline_reader_open(cmdline_reader cr);
int cmdline_reader_is_eof(cmdline_reader cr);
char cmdline_reader_get_char(cmdline_reader cr);
void cmdline_reader_close(cmdline_reader cr);

#endif //INSTRUMENTATION_CMDLINE_READER_H
