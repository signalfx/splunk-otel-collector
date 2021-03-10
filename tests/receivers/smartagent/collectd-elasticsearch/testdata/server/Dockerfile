FROM docker.elastic.co/elasticsearch/elasticsearch:7.5.1

ENV ELASTIC_PASSWORD="testing123"
ENV discovery.type="single-node"
ENV ES_JAVA_OPTS="-Xms128m -Xmx128m"

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["eswrapper"]
