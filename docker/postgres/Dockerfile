FROM postgres:17-alpine

ENV POSTGRES_USER=root
ENV POSTGRES_PASSWORD=otel
ENV POSTGRES_DB=otel

COPY initPostgresqlDB.sql /docker-entrypoint-initdb.d/
