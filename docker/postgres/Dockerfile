FROM postgres:18-alpine

ENV POSTGRES_USER=root
ENV POSTGRES_PASSWORD=otel
ENV POSTGRES_DB=otel

COPY initPostgresqlDB.sql /docker-entrypoint-initdb.d/
