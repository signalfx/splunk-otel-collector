ARG COUCHBASE_VERSION=7.6.1
FROM couchbase/server:${COUCHBASE_VERSION}
COPY entrypoint.sh /config-entrypoint.sh
EXPOSE 8091
ENTRYPOINT ["/config-entrypoint.sh"]
