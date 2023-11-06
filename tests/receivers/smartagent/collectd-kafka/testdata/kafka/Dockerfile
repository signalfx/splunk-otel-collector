FROM bitnami/kafka:3.6.0

ENV JMX_PORT=7099
EXPOSE 7099

ARG KAFKA_VERSION=3.6.0
ENV KAFKA_VERSION=$KAFKA_VERSION
ENV KAFKA_BIN="/opt/bitnami/kafka/bin"  

ADD scripts/* scripts/
CMD ["bash", "scripts/run.sh"]
