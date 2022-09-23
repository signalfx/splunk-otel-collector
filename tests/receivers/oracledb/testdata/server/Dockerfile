FROM container-registry.oracle.com/database/express:21.3.0-xe

ENV ORACLE_PWD=mysecurepassword

COPY initOracleDB.sql $ORACLE_BASE/scripts/startup
