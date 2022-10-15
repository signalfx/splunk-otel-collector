
/* The alter session command is required to enable user creation in an Oracle docker container
   This command shouldn't be used outside of test environments. */
alter session set "_ORACLE_SCRIPT"=true;
CREATE USER OTEL IDENTIFIED BY password;
GRANT CREATE SESSION TO OTEL;
GRANT SELECT ON V_$SESSION TO OTEL;
GRANT SELECT ON V_$SYSSTAT TO OTEL;
GRANT SELECT ON V_$RESOURCE_LIMIT TO OTEL;
GRANT SELECT ON DBA_TABLESPACES TO OTEL;
GRANT SELECT ON DBA_DATA_FILES TO OTEL;
