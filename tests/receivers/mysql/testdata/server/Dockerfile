FROM mysql:8.2.0

RUN chown -R mysql:root /var/lib/mysql/

ENV MYSQL_ROOT_PASSWORD='testuser'

COPY initMySqlDB.sql /docker-entrypoint-initdb.d/

EXPOSE 3306