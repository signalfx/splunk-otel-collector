FROM felfert/wix:latest

USER root
RUN apt-get update -y
RUN apt-get install -y curl unzip

USER wix

COPY ./build.sh /work/build.sh

ENV OUTPUT_DIR=project/dist
ENV SMART_AGENT_RELEASE=5.9.1

# `wix` user cannot write to anything outside /work, so run w/ `-u 0` if necessary (circle configuration)
ENTRYPOINT ["bash", "-c", "su wix -c \"/work/build.sh $0 $@ --output /work/build/stage --smart-agent $SMART_AGENT_RELEASE \" && cp /work/build/stage/*.msi $OUTPUT_DIR/"]
