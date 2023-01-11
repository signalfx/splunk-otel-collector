ARG DOCKER_REPO=docker.io
FROM ${DOCKER_REPO}/debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY testdata/instrumentation.conf /libsplunk/testdata/instrumentation.conf
COPY testdata/instrumentation-svcname.conf /libsplunk/testdata/instrumentation-svcname.conf
COPY install/instrumentation.conf /libsplunk/install/instrumentation.conf
COPY Makefile /libsplunk/Makefile
