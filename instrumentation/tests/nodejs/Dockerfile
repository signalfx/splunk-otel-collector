ARG BASE=busybox
FROM $BASE
ARG ARCH=amd64

COPY index.js .

RUN mkdir -p /etc/splunk/zeroconfig

COPY zeroconfig.conf /etc/splunk/zeroconfig/node.conf

CMD node index.js

COPY libotelinject_${ARCH}.so /usr/lib/splunk-instrumentation/libotelinject_${ARCH}.so

RUN echo /usr/lib/splunk-instrumentation/libotelinject_${ARCH}.so >> /etc/ld.so.preload
