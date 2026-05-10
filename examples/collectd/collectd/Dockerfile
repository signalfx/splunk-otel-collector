FROM ubuntu:26.04

RUN apt-get update && apt-get install -y collectd && apt-get clean

CMD ["/usr/sbin/collectd", "-f"]