FROM felfert/wix:latest

COPY ./build.sh /work/build.sh

ENV OUTPUT_DIR=project/dist

# `wix` user cannot write to anything outside /work, so run w/ `-u 0` if necessary (circle configuration)
ENTRYPOINT ["bash", "-c", "su wix -c \"/work/build.sh $0 $@ --output /work/build/stage \" && cp /work/build/stage/*.msi $OUTPUT_DIR/"]
