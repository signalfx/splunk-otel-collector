FROM alpine:3.14.2 as certs
RUN apk --update add ca-certificates

FROM alpine:3.14.2 AS otelcol
COPY otelcol /
COPY translatesfx /
COPY migratecheckpoint /
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see https://github.com/open-telemetry/opentelemetry-collector/issues/1317)
RUN chmod 755 /otelcol
RUN chmod 755 /translatesfx
RUN chmod 755 /migratecheckpoint
RUN echo "splunk-otel-collector:x:999:999::/:" > /etc_passwd
# create base dirs since we cannot chown in scratch image except via COPY
RUN mkdir -p /otel/collector /splunk-otel-collector

FROM alpine:3.14.2 AS smartagent
ARG SMART_AGENT_RELEASE

COPY --from=otelcol /etc_passwd /etc_passwd
RUN cat /etc_passwd >> /etc/passwd
COPY --from=otelcol --chown=999 /splunk-otel-collector /usr/lib/splunk-otel-collector
USER splunk-otel-collector
RUN SMART_AGENT=signalfx-agent-${SMART_AGENT_RELEASE#v}.tar.gz && \
    URL=https://github.com/signalfx/signalfx-agent/releases/download/${SMART_AGENT_RELEASE}/$SMART_AGENT && \
    cd /usr/lib/splunk-otel-collector && wget "$URL" && tar -xzf $SMART_AGENT && mv signalfx-agent agent-bundle && \
    # Absolute path of interpreter in smart agent dir is set in dependent binaries
    # requiring the interpreter location not to change.
    /usr/lib/splunk-otel-collector/agent-bundle/bin/patch-interpreter /usr/lib/splunk-otel-collector/agent-bundle && \
    rm -f /usr/lib/splunk-otel-collector/agent-bundle/bin/signalfx-agent \
          /usr/lib/splunk-otel-collector/agent-bundle/bin/agent-status \
          $SMART_AGENT
RUN find /usr/lib/splunk-otel-collector/agent-bundle -name "*.key" -delete -or -name "*.pem" -delete

FROM debian:10.11 as journal
RUN apt update
RUN apt install -y systemd

FROM scratch
ENV SPLUNK_BUNDLE_DIR=/usr/lib/splunk-otel-collector/agent-bundle
ENV SPLUNK_COLLECTD_DIR=${SPLUNK_BUNDLE_DIR}/run/collectd
ENV PATH=${PATH}:${SPLUNK_BUNDLE_DIR}/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /etc_passwd /etc/passwd
COPY --from=otelcol --chown=999 /otelcol /
COPY --from=otelcol --chown=999 /translatesfx /
COPY --from=otelcol --chown=999 /migratecheckpoint /
COPY --from=otelcol --chown=999 /otel /etc/otel
COPY --from=otelcol --chown=999 /otel/collector /etc/otel/collector
COPY --from=smartagent --chown=999 /usr/lib/splunk-otel-collector /usr/lib/splunk-otel-collector

COPY --chown=999 config/collector/gateway_config.yaml /etc/otel/collector/gateway_config.yaml
COPY --chown=999 config/collector/otlp_config_linux.yaml /etc/otel/collector/otlp_config_linux.yaml
COPY --chown=999 config/collector/agent_config.yaml /etc/otel/collector/agent_config.yaml
COPY --chown=999 config/collector/fargate_config.yaml /etc/otel/collector/fargate_config.yaml
COPY --chown=999 config/collector/ecs_ec2_config.yaml /etc/otel/collector/ecs_ec2_config.yaml

# Below files are required by journald receiver which are referenced from https://altlinux.pkgs.org/sisyphus/classic-x86_64/systemd-utils-249.5-alt2.x86_64.rpm.html
COPY --from=journal /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=journal /lib/systemd/libsystemd-shared-241.so /lib/systemd/libsystemd-shared-241.so
COPY --from=journal /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=journal /lib/x86_64-linux-gnu/librt.so.1 /lib/x86_64-linux-gnu/librt.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libcap.so.2 /lib/x86_64-linux-gnu/libcap.so.2
COPY --from=journal /usr/lib/x86_64-linux-gnu/libacl.so.1 /usr/lib/x86_64-linux-gnu/libacl.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libcryptsetup.so.12 /lib/x86_64-linux-gnu/libcryptsetup.so.12
COPY --from=journal /lib/x86_64-linux-gnu/libgcrypt.so.20 /lib/x86_64-linux-gnu/libgcrypt.so.20
COPY --from=journal /usr/lib/x86_64-linux-gnu/libip4tc.so.0 /usr/lib/x86_64-linux-gnu/libip4tc.so.0
COPY --from=journal /usr/lib/x86_64-linux-gnu/libkmod.so.2 /usr/lib/x86_64-linux-gnu/libkmod.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libmount.so.1 /lib/x86_64-linux-gnu/libmount.so.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libseccomp.so.2 /usr/lib/x86_64-linux-gnu/libseccomp.so.2
COPY --from=journal /lib/x86_64-linux-gnu/libselinux.so.1 /lib/x86_64-linux-gnu/libselinux.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libidn.so.11 /lib/x86_64-linux-gnu/libidn.so.11
COPY --from=journal /lib/x86_64-linux-gnu/liblzma.so.5 /lib/x86_64-linux-gnu/liblzma.so.5
COPY --from=journal /usr/lib/x86_64-linux-gnu/liblz4.so.1 /usr/lib/x86_64-linux-gnu/liblz4.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libblkid.so.1 /lib/x86_64-linux-gnu/libblkid.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=journal /usr/lib/x86_64-linux-gnu/libattr.so.1 /usr/lib/x86_64-linux-gnu/libattr.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libuuid.so.1 /lib/x86_64-linux-gnu/libuuid.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libdevmapper.so.1.02.1 /lib/x86_64-linux-gnu/libdevmapper.so.1.02.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libssl.so.1.1 /usr/lib/x86_64-linux-gnu/libssl.so.1.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1
COPY --from=journal /usr/lib/x86_64-linux-gnu/libargon2.so.1 /usr/lib/x86_64-linux-gnu/libargon2.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=journal /usr/lib/x86_64-linux-gnu/libjson-c.so.3 /usr/lib/x86_64-linux-gnu/libjson-c.so.3
COPY --from=journal /lib/x86_64-linux-gnu/libgpg-error.so.0 /lib/x86_64-linux-gnu/libgpg-error.so.0
COPY --from=journal /lib/x86_64-linux-gnu/libpcre.so.3 /lib/x86_64-linux-gnu/libpcre.so.3
COPY --from=journal /lib/x86_64-linux-gnu/libudev.so.1 /lib/x86_64-linux-gnu/libudev.so.1
COPY --from=journal /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=journal /bin/journalctl /bin/journalctl

USER splunk-otel-collector
ENTRYPOINT ["/otelcol"]
EXPOSE 13133 14250 14268 4317 4318 6060 8888 9411 9443 9080
