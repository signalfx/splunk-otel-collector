FROM mysql:9.3
ENV MYSQL_DATABASE testdb
ENV MYSQL_USER testuser
ENV MYSQL_PASSWORD testpass
ENV MYSQL_ROOT_PASSWORD testpass
COPY init.sql /docker-entrypoint-initdb.d
