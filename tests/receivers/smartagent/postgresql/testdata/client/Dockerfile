ARG POSTGRES_VERSION=13-alpine
FROM postgres:${POSTGRES_VERSION}

COPY requests.sh /usr/local/bin/requests.sh

CMD ["requests.sh"]
