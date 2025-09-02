#
# Salt Stack Salt Dev Container
#

FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV SALT_VERSION=latest
ENV SALT_LINT_VERSION=0.9.2

RUN apt-get update && apt-get upgrade -y -o DPkg::Options::=--force-confold
RUN apt-get install -y python3-pip curl

RUN mkdir -p /etc/apt/keyrings
RUN curl -fsSL https://packages.broadcom.com/artifactory/api/security/keypair/SaltProjectKey/public | tee /etc/apt/keyrings/salt-archive-keyring.pgp
RUN curl -fsSL https://github.com/saltstack/salt-install-guide/releases/${SALT_VERSION}/download/salt.sources | tee /etc/apt/sources.list.d/salt.sources
RUN apt-get update
RUN apt-get install -y salt-minion

RUN pipx install salt-lint==${SALT_LINT_VERSION}

RUN sed -i "s|#file_client:.*|file_client: local|" /etc/salt/minion

COPY ./Makefile /Makefile
COPY ./splunk-otel-collector /srv/salt/splunk-otel-collector
COPY ./templates /srv/salt/templates
COPY ./pillar.example /srv/pillar/splunk-otel-collector.sls
COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /srv/salt/splunk-otel-collector

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
