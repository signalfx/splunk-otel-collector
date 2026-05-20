ARG BASE=busybox
FROM $BASE

COPY Main.java .

RUN javac Main.java && \
    jar cfe Main.jar Main Main.class

COPY injector.conf /etc/opentelemetry/injector.conf

CMD java -jar Main.jar

ENV OTEL_SERVICE_NAME=iknowmyownservicename

ENV ANOTHER_VAR=foo
ENV SPLUNK_METRICS_ENABLED=true

COPY libotelinject.so /usr/lib/splunk-instrumentation/libotelinject.so

RUN echo /usr/lib/splunk-instrumentation/libotelinject.so >> /etc/ld.so.preload
