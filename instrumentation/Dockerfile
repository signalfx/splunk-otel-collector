FROM debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY config/auto-instr.conf /libsplunk/testdata/splunk.conf
COPY Makefile /libsplunk/Makefile
