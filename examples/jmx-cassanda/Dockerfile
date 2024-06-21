FROM cassandra:latest

ENV LOCAL_JMX no
COPY jmxremote.access /etc/cassandra/jmxremote.access
COPY jmxremote.password /etc/cassandra/jmxremote.password
RUN chown cassandra:cassandra /etc/cassandra/jmxremote.*
RUN chmod 400 /etc/cassandra/jmxremote.*

# cassandra-env is small enough to read into memory, KISS, no (s)ed etc.
RUN echo "`echo 'JMX_LOCAL="no"'; echo 'JVM_OPTS="$JVM_OPTS -Dcom.sun.management.jmxremote.access.file=/etc/cassandra/jmxremote.access"'; cat /etc/cassandra/cassandra-env.sh`" > /etc/cassandra/cassandra-env.sh

