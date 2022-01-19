#
# Salt Stack Salt Dev Container
#

FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get upgrade -y -o DPkg::Options::=--force-confold
RUN apt-get install -y software-properties-common ca-certificates wget curl apt-transport-https python3-pip vim

RUN curl -L https://repo.saltproject.io/py3/ubuntu/20.04/amd64/latest/SALTSTACK-GPG-KEY.pub | apt-key add -
RUN echo 'deb http://repo.saltproject.io/py3/ubuntu/20.04/amd64/latest focal main' > /etc/apt/sources.list.d/saltstack.list && \
    apt-get update && \
    apt-get install -y salt-minion

RUN pip3 install salt-lint==0.8.0

RUN sed -i "s|#file_client:.*|file_client: local|" /etc/salt/minion

COPY ./Makefile /Makefile
COPY ./splunk-otel-collector /srv/salt/splunk-otel-collector
COPY ./templates /srv/salt/templates
COPY ./pillar.example /srv/pillar/splunk-otel-collector.sls
COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /srv/salt/splunk-otel-collector

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
