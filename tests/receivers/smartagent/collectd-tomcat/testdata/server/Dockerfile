ARG TOMCAT_VERSION=latest
FROM tomcat:${TOMCAT_VERSION}
ENV JAVA_TOOL_OPTIONS "-Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.port=5000  -Dcom.sun.management.jmxremote.rmi.port=5000 -Dcom.sun.management.jmxremote.host=0.0.0.0  -Djava.rmi.server.hostname=0.0.0.0"
CMD ["catalina.sh", "run"]