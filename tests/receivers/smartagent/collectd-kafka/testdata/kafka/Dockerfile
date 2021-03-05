FROM ubuntu:16.04

ENV JMX_PORT=7099
EXPOSE 7099

RUN apt-get update
RUN apt-get install -y wget
RUN apt-get install -y default-jre

ARG KAFKA_VERSION=1.0.1
ENV KAFKA_VERSION=$KAFKA_VERSION
ENV KAFKA_BIN="/opt/kafka_2.11-$KAFKA_VERSION/bin"

RUN cd /opt/ && wget https://archive.apache.org/dist/kafka/"$KAFKA_VERSION"/kafka_2.11-"$KAFKA_VERSION".tgz && \
    tar -zxf kafka_2.11-"$KAFKA_VERSION".tgz && cd kafka_2.11-"$KAFKA_VERSION"/
ADD scripts/* scripts/
CMD ["bash", "scripts/run.sh"]
