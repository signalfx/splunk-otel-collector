FROM --platform=$BUILDPLATFORM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /source

# copy csproj and restore as distinct layers
COPY *.csproj .
RUN dotnet restore --use-current-runtime

# copy and publish app and libraries
COPY . .
RUN dotnet publish --use-current-runtime --no-restore -o /app

# final stage/image
FROM mcr.microsoft.com/dotnet/runtime:8.0
WORKDIR /app
COPY --from=build /app .

COPY zeroconfig.conf /etc/splunk/zeroconfig/dotnet.conf

COPY libsplunk.so /usr/lib/splunk-instrumentation/libsplunk.so
ENV OTEL_SERVICE_NAME iknowmyownservicename

ENV ANOTHER_VAR foo

RUN echo /usr/lib/splunk-instrumentation/libsplunk.so >> /etc/ld.so.preload

CMD dotnet dotnetapp.dll