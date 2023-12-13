FROM mysql:8.2.0

RUN chown -R mysql:root /var/lib/mysql/

ENV MYSQL_DATABASE="otel"
ENV MYSQL_USER="otelu"
ENV MYSQL_PASSWORD="otelp"
ENV MYSQL_ROOT_PASSWORD='testuser'

RUN microdnf update && microdnf -y install vim

RUN cp initMysqlDB.sql /docker-entrypoint-initdb.d

EXPOSE 3306

