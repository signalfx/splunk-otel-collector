FROM cassandra:5.0

# Configures Cassandra with remote JMX with username/password of
# cassandra/cassandra

ENV LOCAL_JMX=no

RUN echo 'cassandra readonly' > /etc/cassandra/jmxremote.access &&\
    echo 'cassandra cassandra' > /etc/cassandra/jmxremote.password &&\
	echo 'JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.access.file=/etc/cassandra/jmxremote.access"' >> /etc/cassandra/cassandra-env.sh &&\
    chmod 0400 /etc/cassandra/jmxremote* &&\
    chown cassandra:cassandra /etc/cassandra/jmxremote.* &&\
    chown cassandra:cassandra /etc/cassandra/cassandra-env.sh
