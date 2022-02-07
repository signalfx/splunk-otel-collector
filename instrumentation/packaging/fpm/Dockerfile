FROM debian:10

VOLUME /instrumentation
WORKDIR /instrumentation

ENV PACKAGE="deb"
ENV VERSION=""
ENV ARCH="amd64"
ENV OUTPUT_DIR="/instrumentation/dist/"

COPY install-deps.sh /install-deps.sh

RUN /install-deps.sh

CMD ./packaging/fpm/$PACKAGE/build.sh "$VERSION" "$ARCH" "$OUTPUT_DIR"
