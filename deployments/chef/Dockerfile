FROM ruby:2.6-buster

RUN apt-get update &&\
    apt-get install -yq ca-certificates procps systemd apt-transport-https libcap2-bin curl gnupg

COPY ./ /splunk-otel-collector
WORKDIR /splunk-otel-collector

RUN gem install bundler && bundle install
