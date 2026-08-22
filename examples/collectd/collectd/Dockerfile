FROM ubuntu:26.04

RUN apt-get update && apt-get install -y collectd && apt-get clean

USER nobody:nogroup

CMD ["/usr/sbin/collectd", "-f"]
