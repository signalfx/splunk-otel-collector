FROM mysql:9.4
ENV MYSQL_DATABASE testdb
ENV MYSQL_USER testuser
ENV MYSQL_PASSWORD testpass
ENV MYSQL_ROOT_PASSWORD testpass
COPY init.sql /docker-entrypoint-initdb.d
