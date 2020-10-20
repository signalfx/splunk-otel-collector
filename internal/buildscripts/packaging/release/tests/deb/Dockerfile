# A debian9 image with systemd enabled.  Must be run with:
# `-d --privileged -v /sys/fs/cgroup:/sys/fs/cgroup:ro` flags
FROM debian:9

RUN apt-get update &&\
    apt-get install -yq apt-transport-https ca-certificates curl procps systemd

ENV container docker
RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i = \
    "systemd-tmpfiles-setup.service" ] || rm -f $i; done); \
    rm -f /lib/systemd/system/multi-user.target.wants/*;\
    rm -f /lib/systemd/system/local-fs.target.wants/*; \
    rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
    rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
    rm -f /lib/systemd/system/anaconda.target.wants/*;

RUN systemctl set-default multi-user.target
ENV init /lib/systemd/systemd

ARG STAGE=release
RUN curl -sSL https://splunk.jfrog.io/splunk/otel-collector-deb/splunk-B3CD4420.gpg > /etc/apt/trusted.gpg.d/splunk.gpg
RUN echo "deb https://splunk.jfrog.io/splunk/otel-collector-deb $STAGE main" > /etc/apt/sources.list.d/splunk-otel-collector.list

VOLUME [ "/sys/fs/cgroup" ]

ENTRYPOINT ["/lib/systemd/systemd"]
