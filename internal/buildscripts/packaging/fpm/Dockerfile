FROM debian:10

VOLUME /repo
WORKDIR /repo

ENV PACKAGE="deb"
ENV VERSION=""
ENV ARCH="amd64"
ENV OUTPUT_DIR="/repo/dist/"

COPY install-deps.sh Gemfile Gemfile.lock /

RUN /install-deps.sh

CMD ./internal/buildscripts/packaging/fpm/$PACKAGE/build.sh "$VERSION" "$ARCH" "$OUTPUT_DIR" "$SMART_AGENT_RELEASE"
