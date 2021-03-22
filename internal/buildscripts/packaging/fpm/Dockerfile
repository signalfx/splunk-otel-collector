FROM debian:9

RUN apt-get update && \
    apt-get install -y ruby ruby-dev rubygems build-essential git rpm sudo curl jq

RUN gem install --no-document fpm -v 1.11.0

VOLUME /repo
WORKDIR /repo

ENV PACKAGE="deb"
ENV VERSION=""
ENV ARCH="amd64"
ENV OUTPUT_DIR="/repo/dist/"
ENV SMART_AGENT_RELEASE="v5.9.1"

CMD ./internal/buildscripts/packaging/fpm/$PACKAGE/build.sh "$VERSION" "$ARCH" "$OUTPUT_DIR" "$SMART_AGENT_RELEASE"
