FROM --platform=$BUILDPLATFORM mcr.microsoft.com/dotnet/sdk:8.0 AS build
ARG TARGETARCH
WORKDIR /source

# copy csproj and restore as distinct layers
COPY *.csproj .
RUN dotnet restore -a ${TARGETARCH}

# copy and publish app and libraries
COPY . .
RUN dotnet publish -a $TARGETARCH --no-restore -o /app

# final stage/image
FROM mcr.microsoft.com/dotnet/runtime:8.0
ARG TARGETARCH
WORKDIR /app
COPY --from=build /app .

COPY zeroconfig.conf /etc/splunk/zeroconfig/dotnet.conf

COPY libotelinject_${TARGETARCH}.so /usr/lib/splunk-instrumentation/libotelinject_${TARGETARCH}.so
ENV OTEL_SERVICE_NAME iknowmyownservicename

ENV ANOTHER_VAR foo

RUN echo /usr/lib/splunk-instrumentation/libotelinject_${TARGETARCH}.so >> /etc/ld.so.preload

CMD dotnet dotnetapp.dll