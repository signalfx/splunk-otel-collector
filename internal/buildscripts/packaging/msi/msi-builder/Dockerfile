FROM felfert/wix:latest

USER root
RUN apt-get update -y
RUN apt-get install -y curl unzip

COPY ./build.sh /work/build.sh
COPY ./docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod a+x /work/build.sh /docker-entrypoint.sh

ENV OUTPUT_DIR=project/dist
ENV SMART_AGENT_RELEASE=5.10.2

ENTRYPOINT ["/docker-entrypoint.sh"]
