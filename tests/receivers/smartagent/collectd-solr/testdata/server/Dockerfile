FROM bitnami/solr:9.2.1

ENV SOLR_JETTY_HOST 0.0.0.0
ENTRYPOINT ["bash", "-c"]
CMD ["solr start -c -e techproducts && tail -f /dev/null"]

