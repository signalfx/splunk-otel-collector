#!/bin/bash
set -euo pipefail

TOMCAT_HOME="/usr/local/tomcat"
JAVA_HOME="/opt/java/openjdk"

wget --no-check-certificate -nv -O ${TOMCAT_HOME}/webapps/sample.war https://tomcat.apache.org/tomcat-9.0-doc/appdev/sample/sample.war

cat <<EOH > /etc/systemd/system/tomcat.service
[Unit]
Description=Apache Tomcat Web Application Container
After=network.target

[Service]
Type=forking

Environment="JAVA_HOME=${JAVA_HOME}"
Environment="CATALINA_PID=${TOMCAT_HOME}/temp/tomcat.pid"
Environment="CATALINA_HOME=${TOMCAT_HOME}"
Environment="CATALINA_BASE=${TOMCAT_HOME}"
Environment="CATALINA_OPTS=-Xms512M -Xmx1024M -server -XX:+UseParallelGC"
Environment="JAVA_OPTS=-Djava.awt.headless=true"

ExecStart=${TOMCAT_HOME}/bin/startup.sh
ExecStop=${TOMCAT_HOME}/bin/shutdown.sh

User=tomcat
Group=tomcat
UMask=0007
RestartSec=10
Restart=always

[Install]
WantedBy=multi-user.target
EOH

useradd -r -m -U -d $TOMCAT_HOME -s /bin/false tomcat
chown -R tomcat:tomcat $TOMCAT_HOME
