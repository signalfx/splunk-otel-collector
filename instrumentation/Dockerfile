FROM debian:11

RUN apt-get update && \
    apt-get install -y build-essential

WORKDIR /libsplunk

COPY src /libsplunk/src
COPY config.example.txt /libsplunk/config.example.txt
COPY Makefile /libsplunk/Makefile
