FROM ruby:3.3-bookworm

RUN apt-get update &&\
    apt-get install -yq ca-certificates procps systemd apt-transport-https libcap2-bin curl gnupg

WORKDIR /splunk-otel-collector

COPY Gemfile /splunk-otel-collector/

RUN bundle install

COPY ./ /splunk-otel-collector
