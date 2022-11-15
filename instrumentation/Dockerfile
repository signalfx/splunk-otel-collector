FROM debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY testdata/instrumentation.conf /libsplunk/testdata/instrumentation.conf
COPY install/instrumentation.conf /libsplunk/install/instrumentation.conf
COPY Makefile /libsplunk/Makefile
