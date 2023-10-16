FROM bitnami/kafka:1.0.1-r2

ENV JMX_PORT=7099
EXPOSE 7099

ARG KAFKA_VERSION=1.0.1
ENV KAFKA_VERSION=$KAFKA_VERSION
ENV KAFKA_BIN="/opt/bitnami/kafka/bin"

ADD scripts/* scripts/
CMD ["bash", "scripts/run.sh"]
