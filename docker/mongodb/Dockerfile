ARG MONGODB_VERSION=4.0
FROM mongo:${MONGODB_VERSION}

ENV MONGO_INITDB_ROOT_USERNAME="admin" 
ENV MONGO_INITDB_ROOT_PASSWORD="pass"

COPY scripts /home/mongodb/scripts
COPY mongod.conf /home/mongodb

WORKDIR /home/mongodb

RUN chmod +x /home/mongodb/scripts/*.sh

CMD ["/bin/bash", "/home/mongodb/scripts/run.sh"]
