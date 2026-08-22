# This Dockerfile is used by the integration example and installs additional software to try out the Splunk_TA_nix addon.
FROM debian:13

RUN mkdir -p /bin

COPY --chmod=755 bin/otelcol_linux_amd64 /bin/otelcol

# Required by Splunk_TA_nix addon
RUN apt-get update && apt-get install -y net-tools lastlog2 ntpsec-ntpdate lsof sysstat auditd

ENTRYPOINT ["/bin/otelcol"]
