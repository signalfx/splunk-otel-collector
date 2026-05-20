ARG BASE=busybox
FROM $BASE

COPY index.js .

RUN mkdir -p /etc/opentelemetry

COPY injector.conf /etc/opentelemetry/injector.conf

CMD node index.js

ENV SPLUNK_METRICS_ENABLED=true

COPY libotelinject.so /usr/lib/splunk-instrumentation/libotelinject.so

RUN echo /usr/lib/splunk-instrumentation/libotelinject.so >> /etc/ld.so.preload
