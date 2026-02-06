ARG BASE=busybox
FROM $BASE
ARG ARCH=amd64

COPY Main.java .

RUN javac Main.java && \
    jar cfe Main.jar Main Main.class

COPY zeroconfig.conf /etc/splunk/zeroconfig/java.conf

CMD java -jar Main.jar

ENV OTEL_SERVICE_NAME=iknowmyownservicename

ENV ANOTHER_VAR=foo

COPY libotelinject_${ARCH}.so /usr/lib/splunk-instrumentation/libotelinject_${ARCH}.so

RUN echo /usr/lib/splunk-instrumentation/libotelinject_${ARCH}.so >> /etc/ld.so.preload
