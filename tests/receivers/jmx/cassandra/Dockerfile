FROM eclipse-temurin:17-jre-alpine AS java
FROM otelcol:latest
COPY --from=java / /java/
ENV JAVA_HOME=/java PATH="/java/bin:${PATH}"
