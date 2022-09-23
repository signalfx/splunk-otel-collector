
/* The alter session command is required to enable user creation in an Oracle docker container
   This command shouldn't be used outside of test environments. */
alter session set "_ORACLE_SCRIPT"=true;
CREATE USER OTEL IDENTIFIED BY password;
GRANT CREATE SESSION TO OTEL;
GRANT SELECT ON V_$SQLSTATS TO OTEL;
GRANT SELECT ON V_$SESSION TO OTEL;
GRANT SELECT ON V_$SESSTAT TO OTEL;
GRANT SELECT ON V_$STATNAME TO OTEL;
GRANT SELECT ON V_$SESSMETRIC TO OTEL;
GRANT SELECT ON V_$SESSION_LONGOPS TO OTEL;
GRANT SELECT ON V_$RESOURCE_LIMIT TO OTEL;