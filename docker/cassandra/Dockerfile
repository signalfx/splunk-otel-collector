FROM cassandra:3.11

# Configures Cassandra with remote JMX with username/password of
# cassandra/cassandra

ENV LOCAL_JMX=no

RUN echo 'cassandra readonly' > /etc/cassandra/jmxremote.access &&\
    echo 'cassandra cassandra' > /etc/cassandra/jmxremote.password &&\
	chmod 0400 /etc/cassandra/jmxremote*
