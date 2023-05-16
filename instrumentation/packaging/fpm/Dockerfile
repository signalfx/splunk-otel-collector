FROM debian:10

VOLUME /repo
WORKDIR /repo

COPY install-deps.sh Gemfile Gemfile.lock /

RUN /install-deps.sh

RUN git config --global --add safe.directory /repo
