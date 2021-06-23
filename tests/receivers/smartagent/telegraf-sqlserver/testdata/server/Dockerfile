FROM mcr.microsoft.com/mssql/server:2019-latest

ENV ACCEPT_EULA=Y
ENV SA_PASSWORD=Password!
ENV HOSTNAME=sql.example.com

USER root
RUN mkdir -p /home/mssql/ssl
RUN chown -R mssql /home/mssql
COPY mssql.conf /var/opt/mssql/mssql.conf

USER mssql
RUN openssl req -x509 -nodes -newkey rsa:2048 -subj '/CN=sql.example.com' -keyout /home/mssql/ssl/mssql.key -out /home/mssql/ssl/mssql.pem -days 365
RUN chmod 400 /home/mssql/ssl/mssql.pem
RUN chmod 400 /home/mssql/ssl/mssql.key
