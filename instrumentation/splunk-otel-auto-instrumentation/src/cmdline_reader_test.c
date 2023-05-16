#include "cmdline_reader.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

struct cmdline_reader_impl {
    int i;
    unsigned long size;
    char *cmdline;
};

cmdline_reader new_cmdline_reader() {
    // It appears that just having the
    // void __attribute__((constructor)) splunk_instrumentation_enter()
    // function in the executable causes it to run during tests, so we
    // create a cmdline_reader here.
    return new_test_cmdline_reader("", 0);
}

cmdline_reader new_test_cmdline_reader(char *cmdline, int size) {
    cmdline_reader cr = malloc(sizeof(*cr));
    cr->i = 0;

    cr->cmdline = malloc(size);
    memcpy(cr->cmdline, cmdline, size);

    cr->size = size;
    return cr;
}

void cmdline_reader_open(cmdline_reader cr) {
}

int cmdline_reader_is_eof(cmdline_reader cr) {
    return cr->i >= cr->size;
}

char cmdline_reader_get_char(cmdline_reader cr) {
    return cr->cmdline[cr->i++];
}

void cmdline_reader_close(cmdline_reader cr) {
    free(cr->cmdline);
}
