FROM mcr.microsoft.com/mssql-tools

COPY create_user.sql /usr/local/create_user.sql

CMD ["bash", "-c", "sleep 10 && /opt/mssql-tools/bin/sqlcmd -S \"tcp:sql-server,1433\" -U sa -P \"Password!\" -i /usr/local/create_user.sql"]
