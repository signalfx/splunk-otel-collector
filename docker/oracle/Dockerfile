FROM gvenzl/oracle-xe:21-slim-faststart

ENV ORACLE_PASSWORD=mysecurepassword

COPY initOracleDB.sql /container-entrypoint-initdb.d/startup.sql

HEALTHCHECK CMD "$ORACLE_BASE/healthcheck.sh"