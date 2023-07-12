FROM eclipse-temurin:8-jdk

WORKDIR /root

# install dependencies
RUN apt-get update \
    && apt-get install -y \
    curl \
    openssh-server \
    tar

RUN systemctl enable ssh.service

## set environment variables for hadoop
ENV JAVA_HOME=/opt/java/openjdk

# setup ssh
RUN ssh-keygen -t rsa -f /root/.ssh/id_rsa -P '' \
    && cp /root/.ssh/id_rsa.pub /root/.ssh/authorized_keys

# mkdir for hadoop files
RUN mkdir -p /usr/local/hadoop

# specify the hadoop verison to use
ARG HADOOP_VER=2.10.1

# curl down hadoop and untar it into place
RUN curl -Lks -o hadoop.tar.gz https://dlcdn.apache.org/hadoop/common/hadoop-${HADOOP_VER}/hadoop-${HADOOP_VER}.tar.gz \
    && tar -xf hadoop.tar.gz -C /usr/local/hadoop --strip-components 1 \
    && rm hadoop.tar.gz

# set environment variables for hadoop path
ENV LD_LIBRARY_PATH=/usr/local/hadoop/lib/native:$LD_LIBRARY_PATH
ENV PATH=$PATH:/usr/local/hadoop/bin:/usr/local/hadoop/sbin
ENV HADOOP_HOME=/usr/local/hadoop
ENV HADOOP_CONF_DIR=/usr/local/hadoop/etc/hadoop


# mkdir for hadoop logs and hdfs data
RUN mkdir -p /usr/local/hadoop/logs \
    && mkdir -p /root/hadoop/data/namenode \
    && mkdir -p /root/hadoop/data/datanode

# copy config files
COPY config/* /usr/local/hadoop/etc/hadoop/
COPY config/ssh_config /root/.ssh/config

COPY run_master.sh /usr/local/bin/run_master.sh

EXPOSE 5677 50070 9870 8088 22

CMD [ "sh", "-c", "service ssh start; echo ready; sleep infinity"]
