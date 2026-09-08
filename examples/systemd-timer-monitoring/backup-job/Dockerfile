FROM debian:bookworm

ENV DEBIAN_FRONTEND=noninteractive container=docker

RUN apt-get update && \
    apt-get install -y --no-install-recommends systemd systemd-sysv curl ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Trim units that don't apply/work inside a container, following the
# well-known "systemd in Docker" pattern.
RUN (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ "$i" = systemd-tmpfiles-setup.service ] || rm -f "$i"; done) && \
    rm -f /lib/systemd/system/multi-user.target.wants/* \
          /etc/systemd/system/*.wants/* \
          /lib/systemd/system/local-fs.target.wants/* \
          /lib/systemd/system/sockets.target.wants/*udev* \
          /lib/systemd/system/sockets.target.wants/*initctl* \
          /lib/systemd/system/basic.target.wants/*

COPY backup-job.sh /usr/local/bin/backup-job.sh
COPY report-job-status.sh /usr/local/bin/report-job-status.sh
RUN chmod +x /usr/local/bin/backup-job.sh /usr/local/bin/report-job-status.sh

COPY backup-job.service /etc/systemd/system/backup-job.service
COPY backup-job.timer /etc/systemd/system/backup-job.timer
RUN systemctl enable backup-job.timer

STOPSIGNAL SIGRTMIN+3

CMD ["/lib/systemd/systemd"]
