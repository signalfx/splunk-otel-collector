ARG POSTGRES_VERSION=13-alpine
FROM postgres:${POSTGRES_VERSION}

COPY initdb.d /docker-entrypoint-initdb.d

CMD ["postgres", \
"-c", "shared_preload_libraries=pg_stat_statements", \
"-c", "wal_level=logical", \
"-c", "max_replication_slots=2"]
