FROM bitnamilegacy/spark:4.0.0

EXPOSE 8080

WORKDIR $SPARK_HOME
COPY metrics.properties $SPARK_HOME/conf/metrics.properties
COPY spark-defaults.conf $SPARK_HOME/conf/spark-defaults.conf

CMD ["./bin/spark-class", "org.apache.spark.deploy.master.Master"]
