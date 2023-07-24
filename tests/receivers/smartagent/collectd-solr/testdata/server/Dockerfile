FROM bitnami/solr:latest

ENV SOLR_JETTY_HOST 0.0.0.0
ENTRYPOINT ["bash", "-c"]
CMD ["solr start -c -e techproducts && tail -f /dev/null"]

