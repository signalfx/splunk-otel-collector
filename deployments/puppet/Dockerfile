FROM ruby:3.4.7-bookworm

ENV PATH=$PATH:/opt/puppetlabs/bin:/opt/puppetlabs/pdk/bin

WORKDIR /tmp

RUN wget https://apt.puppetlabs.com/puppet8-release-bookworm.deb
RUN dpkg -i puppet8-release-bookworm.deb
RUN rm *.deb
RUN apt update
RUN apt install -y puppet-agent

RUN wget https://apt.puppet.com/puppet-tools-release-bookworm.deb
RUN dpkg --force-overwrite -i puppet-tools-release-bookworm.deb
RUN apt update
RUN apt install -y pdk
RUN rm -f *.deb

WORKDIR /etc/puppetlabs/code/modules/splunk_otel_collector
COPY ./ ./
RUN pdk update
# Bundler version matches the version used to generate the Gemfile.lock
RUN gem install bundler -v 2.4.13 && bundle install

RUN mkdir -p /root/.config/puppet && \
    echo "---\n\
disabled: true" > /root/.config/puppet/analytics.yml

ENV PATH=/opt/puppetlabs/bin:$PATH
