ARG BASE=busybox
FROM $BASE

COPY Main.java .

RUN javac Main.java && \
    jar cfe Main.jar Main Main.class

COPY zeroconfig.conf /etc/splunk/zeroconfig/java.conf

CMD java -jar Main.jar

ENV OTEL_SERVICE_NAME=iknowmyownservicename

ENV ANOTHER_VAR=foo

COPY libsplunk.so /usr/lib/splunk-instrumentation/libsplunk.so

RUN echo /usr/lib/splunk-instrumentation/libsplunk.so >> /etc/ld.so.preload
