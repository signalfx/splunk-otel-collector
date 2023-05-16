#include "cmdline_reader.h"
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>

struct cmdline_reader_impl {
    FILE *f;
};

cmdline_reader new_cmdline_reader() {
    cmdline_reader cr = malloc(sizeof(cmdline_reader));
    if (cr != NULL) {
        cr->f = NULL;
    }
    return cr;
}

void cmdline_reader_open(cmdline_reader cr) {
    pid_t pid = getpid();
    char fname[1024];
    sprintf(fname, "/proc/%d/cmdline", pid);
    cr->f = fopen(fname, "r");
}

int cmdline_reader_is_eof(cmdline_reader cr) {
    return feof(cr->f) != 0;
}

char cmdline_reader_get_char(cmdline_reader cr) {
    return (char) fgetc(cr->f);
}

void cmdline_reader_close(cmdline_reader cr) {
    if (cr->f != NULL) {
        fclose(cr->f);
    }
    free(cr);
}
