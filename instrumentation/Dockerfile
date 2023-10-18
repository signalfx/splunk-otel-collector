ARG DOCKER_REPO=docker.io
FROM ${DOCKER_REPO}/debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY Makefile /libsplunk/Makefile
