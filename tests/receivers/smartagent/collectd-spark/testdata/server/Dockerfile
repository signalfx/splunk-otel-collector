FROM gettyimages/spark:2.4.1-hadoop-3.0

RUN apt-get update && \
    apt-get install -y netcat procps

EXPOSE 8080

WORKDIR $SPARK_HOME
COPY metrics.properties $SPARK_HOME/conf/metrics.properties
COPY spark-defaults.conf $SPARK_HOME/conf/spark-defaults.conf

CMD ["./bin/spark-class", "org.apache.spark.deploy.master.Master"]
