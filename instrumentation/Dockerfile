ARG DOCKER_REPO=docker.io
FROM ${DOCKER_REPO}/debian:13

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY Makefile /libsplunk/Makefile
