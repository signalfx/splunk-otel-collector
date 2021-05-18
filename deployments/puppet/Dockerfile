FROM ubuntu:16.04

ENV PATH=$PATH:/opt/puppetlabs/bin:/opt/puppetlabs/pdk/bin

WORKDIR /tmp
RUN apt update &&\
    apt install -y make wget vim gcc ruby ruby-dev libxml2 libxml2-dev libxslt1-dev git apt-transport-https ca-certificates g++

RUN wget https://apt.puppetlabs.com/puppet6-release-xenial.deb &&\
    dpkg -i puppet6-release-xenial.deb &&\
    rm *.deb &&\
    apt update &&\
    apt install -y puppet-agent

RUN wget https://apt.puppet.com/puppet-tools-release-xenial.deb && \
    dpkg -i puppet-tools-release-xenial.deb && \
    apt-get update && \
    apt-get install -y pdk && \
    rm -f *.deb

WORKDIR /etc/puppetlabs/code/modules/splunk_otel_collector
COPY ./Rakefile ./Gemfile ./Gemfile.lock ./
RUN gem install bundler && bundle install

RUN mkdir -p /root/.config/puppet && \
    echo "---\n\
disabled: true" > /root/.config/puppet/analytics.yml

ENV PATH=/opt/puppetlabs/bin:$PATH

COPY ./ ./
