FROM eclipse-temurin:8-jdk

ENV JMX_PORT=7099
EXPOSE 7099

ARG KAFKA_VERSION=1.0.1
ENV KAFKA_VERSION=$KAFKA_VERSION
ENV KAFKA_BIN="/opt/kafka_2.11-$KAFKA_VERSION/bin"

RUN cd /opt/ && wget https://archive.apache.org/dist/kafka/"$KAFKA_VERSION"/kafka_2.11-"$KAFKA_VERSION".tgz && \
    tar -zxf kafka_2.11-"$KAFKA_VERSION".tgz && cd kafka_2.11-"$KAFKA_VERSION"/
ADD scripts/* scripts/
CMD ["bash", "scripts/run.sh"]
