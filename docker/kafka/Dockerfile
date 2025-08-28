FROM bitnamilegacy/kafka:3.9.0

ENV JMX_PORT=7099
EXPOSE 7099

ENV KAFKA_BIN="/opt/bitnami/kafka/bin"

ADD scripts/* scripts/
CMD ["bash", "scripts/run.sh"]
