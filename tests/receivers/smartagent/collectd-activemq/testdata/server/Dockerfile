FROM webcenter/activemq:5.14.3

ENV ACTIVEMQ_SUNJMX_START="-Dcom.sun.management.jmxremote.port=1099 -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.password.file=\${ACTIVEMQ_CONFIG_DIR}/jmx.password -Dcom.sun.management.jmxremote.access.file=\${ACTIVEMQ_CONFIG_DIR}/jmx.access"
ENV ACTIVEMQ_JMX_LOGIN=testuser
ENV ACTIVEMQ_JMX_PASSWORD=testing123
ENV ACTIVEMQ_STATIC_TOPICS=testtopic
ENV ACTIVEMQ_STATIC_QUEUES=testqueue

COPY docker-entrypoint.sh /
RUN chmod a+x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["activemq"]
