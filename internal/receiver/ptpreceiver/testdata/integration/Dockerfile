# PTP receiver integration test harness.
#
# Builds an image with LinuxPTP (ptp4l, pmc) and the Go toolchain. run.sh
# starts a "master" ptp4l sidecar container from this same image, then runs
# this image again as the "slave"/test container: its entrypoint starts a
# local ptp4l instance (software timestamping, no PTP-capable NIC required)
# that synchronizes to the sidecar over a shared Docker network, then runs
# the receiver's real integration test against the real pmc/ptp4l tooling.
# See README.md in this directory for how to run it locally.
#
# Built from source: Debian bookworm's packaged linuxptp (3.1.1) predates
# ptp4l's uds_ro_address option, which this receiver's default config relies
# on.
FROM golang:1.26-bookworm

ARG LINUXPTP_VERSION=v4.4

RUN apt-get update \
    && apt-get install -y --no-install-recommends git build-essential \
    && rm -rf /var/lib/apt/lists/* \
    && git clone --depth 1 --branch "${LINUXPTP_VERSION}" https://github.com/richardcochran/linuxptp.git /tmp/linuxptp \
    && make -C /tmp/linuxptp -j"$(nproc)" \
    && make -C /tmp/linuxptp install \
    && rm -rf /tmp/linuxptp

COPY internal/receiver/ptpreceiver/testdata/integration/ptp4l-master.conf /etc/ptp4l/master.conf
COPY internal/receiver/ptpreceiver/testdata/integration/ptp4l-slave.conf /etc/ptp4l/slave.conf
COPY internal/receiver/ptpreceiver/testdata/integration/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /src
ENTRYPOINT ["entrypoint.sh"]
CMD ["go", "test", "-tags=ptp_integration", "-run", "TestPTPReceiverIntegration", "-v", "./internal/receiver/ptpreceiver/..."]
