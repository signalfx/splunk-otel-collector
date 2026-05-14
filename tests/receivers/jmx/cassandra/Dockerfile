FROM eclipse-temurin:17-jdk AS jre-build
RUN jlink \
    --add-modules ALL-MODULE-PATH \
    --strip-debug \
    --no-man-pages \
    --no-header-files \
    --compress=2 \
    --output /jre
COPY packaging/collect-libs.sh /collect-libs.sh
RUN /collect-libs.sh /opt/java-libs /jre/bin /jre/lib

FROM otelcol:latest
COPY --from=jre-build /jre /java/
COPY --from=jre-build /opt/java-libs /
ENV JAVA_HOME=/java PATH="/java/bin:${PATH}"
