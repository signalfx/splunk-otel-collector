FROM node:16

COPY index.js .

RUN mkdir -p /etc/splunk/

COPY zeroconfig.conf /etc/splunk/zeroconfig_node.conf

CMD node index.js

ENV LD_PRELOAD /etc/ld.preload.so/libsplunk.so

COPY libsplunk.so /etc/ld.preload.so/libsplunk.so