ARG COUCHBASE_VERSION=latest
FROM couchbase/server:${COUCHBASE_VERSION}
COPY entrypoint.sh /config-entrypoint.sh
EXPOSE 8091
ENTRYPOINT ["/config-entrypoint.sh"]
