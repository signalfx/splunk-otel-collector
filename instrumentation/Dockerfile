FROM debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY testdata/splunk.conf /libsplunk/testdata/splunk.conf
COPY Makefile /libsplunk/Makefile
