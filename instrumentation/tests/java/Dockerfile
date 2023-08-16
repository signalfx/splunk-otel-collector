FROM eclipse-temurin:11

COPY Main.java .

RUN javac Main.java && \
    jar cfe Main.jar Main Main.class

RUN mkdir -p /etc/splunk/

COPY zeroconfig.conf /etc/splunk/zeroconfig.conf

CMD java -jar Main.jar

COPY libsplunk.so /etc/ld.preload.so/libsplunk.so

ENV LD_PRELOAD /etc/ld.preload.so/libsplunk.so

