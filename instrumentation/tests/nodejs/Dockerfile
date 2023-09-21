ARG BASE=busybox
FROM $BASE

COPY index.js .

RUN mkdir -p /etc/splunk/zeroconfig

COPY zeroconfig.conf /etc/splunk/zeroconfig/node.conf

CMD node index.js

COPY libsplunk.so /usr/lib/splunk-instrumentation/libsplunk.so

RUN echo /usr/lib/splunk-instrumentation/libsplunk.so >> /etc/ld.so.preload
