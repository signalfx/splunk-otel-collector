# The quay.io/signalfx/wix-dev:fleetdm base image is tagged and pushed from fleetdm/wix:latest
# in order to pin the image in case the original image's latest tag is updated or removed.
# https://github.com/fleetdm/fleet/blob/main/LICENSE (MIT Expat)

FROM quay.io/signalfx/wix-dev:fleetdm

USER root
RUN apt-get update -y
RUN apt-get install -y curl unzip

COPY bin/otelcol_windows_amd64.exe /project/bin/otelcol_windows_amd64.exe
COPY bin/translatesfx_windows_amd64.exe /project/bin/translatesfx_windows_amd64.exe
COPY cmd/ /project/cmd
COPY internal/buildscripts/packaging/fpm/ /project/internal/buildscripts/packaging/fpm
COPY internal/buildscripts/packaging/msi/ /project/internal/buildscripts/packaging/msi

RUN cp /project/internal/buildscripts/packaging/msi/msi-builder/docker-entrypoint.sh /docker-entrypoint.sh && \
    chmod a+x /docker-entrypoint.sh

ENV OUTPUT_DIR=/project/dist
ENV SMART_AGENT_RELEASE=latest
ENV VERSION=""

WORKDIR /work

ENTRYPOINT ["/docker-entrypoint.sh"]
