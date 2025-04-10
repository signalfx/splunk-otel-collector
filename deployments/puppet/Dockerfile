FROM ruby:2.6-buster

ENV PATH=$PATH:/opt/puppetlabs/bin:/opt/puppetlabs/pdk/bin

WORKDIR /tmp

RUN wget https://apt.puppetlabs.com/puppet6-release-buster.deb
RUN dpkg -i puppet6-release-buster.deb
RUN rm *.deb
RUN apt update
RUN apt install -y puppet-agent

RUN wget https://apt.puppet.com/puppet6-tools-release-buster.deb
RUN dpkg -i puppet6-tools-release-buster.deb
RUN apt-get update
RUN apt-get install -y pdk
RUN rm -f *.deb

WORKDIR /etc/puppetlabs/code/modules/splunk_otel_collector
COPY ./ ./
RUN gem install bundler -v 2.4.22 && bundle install

RUN mkdir -p /root/.config/puppet && \
    echo "---\n\
disabled: true" > /root/.config/puppet/analytics.yml

ENV PATH=/opt/puppetlabs/bin:$PATH
