ARG BASE=busybox
FROM $BASE

COPY Main.java .

RUN javac Main.java && \
    jar cfe Main.jar Main Main.class

RUN mkdir -p /etc/splunk/

COPY zeroconfig.conf /etc/splunk/zeroconfig_java.conf

CMD java -jar Main.jar

ENV OTEL_SERVICE_NAME iknowmyownservicename

ENV ANOTHER_VAR foo

ENV LD_PRELOAD /etc/ld.preload.so/libsplunk.so

COPY libsplunk.so /etc/ld.preload.so/libsplunk.so

