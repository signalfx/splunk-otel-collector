ARG MONGODB_VERSION=4.0
FROM mongo:${MONGODB_VERSION}

RUN if [ "$MONGO_MAJOR" = 6.0 ] || [ "$MONGO_MAJOR" = 7.0 ]; then \
    apt-get update && apt-get install -y --no-install-recommends bc && apt-get clean; \
    fi

ENV MONGO_INITDB_ROOT_USERNAME="admin" 
ENV MONGO_INITDB_ROOT_PASSWORD="pass"

COPY scripts /home/mongodb/scripts
COPY mongod.conf /home/mongodb

WORKDIR /home/mongodb

RUN chmod +x /home/mongodb/scripts/*.sh

CMD ["/bin/bash", "/home/mongodb/scripts/run.sh"]
