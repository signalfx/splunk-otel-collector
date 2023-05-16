ARG DOCKER_REPO=docker.io
FROM ${DOCKER_REPO}/debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY testdata/instrumentation-default.conf /libsplunk/testdata/instrumentation-default.conf
COPY testdata/instrumentation-options.conf /libsplunk/testdata/instrumentation-options.conf
COPY install/instrumentation.conf /libsplunk/install/instrumentation.conf
COPY Makefile /libsplunk/Makefile
