FROM cassandra:4

# Configures Cassandra with remote JMX with username/password of
# cassandra/cassandra

ENV LOCAL_JMX=no

RUN echo 'cassandra readonly' > /etc/cassandra/jmxremote.access &&\
    echo 'cassandra cassandra' > /etc/cassandra/jmxremote.password &&\
    #    echo 'LOCAL_JMX=no' >> /etc/cassandra/cassandra-env.sh &&\
    #    echo 'JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.authenticate=true"' >> /etc/cassandra/cassandra-env.sh &&\
    #    echo 'JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.password.file=/etc/cassandra/jmxremote.password"' >> /etc/cassandra/cassandra-env.sh &&\
    echo 'JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.access.file=/etc/cassandra/jmxremote.access"' >> /etc/cassandra/cassandra-env.sh &&\
	chmod 0400 /etc/cassandra/jmxremote* &&\
    chown cassandra:cassandra /etc/cassandra/jmxremote.* &&\
    chown cassandra:cassandra /etc/cassandra/cassandra-env.sh


EXPOSE 7000 7001 7199 9042 9160