FROM ubuntu:24.04

RUN apt-get update && apt-get install -y collectd && apt-get clean

CMD ["/usr/sbin/collectd", "-f"]