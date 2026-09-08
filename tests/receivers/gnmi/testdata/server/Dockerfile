# gNMI target for the receiver integration tests.
#
# Builds the OpenConfig fake gNMI agent, which the openconfig/gnmi repository
# ships for exactly this purpose: "implements a gRPC gNMI agent for testing
# against a collector implementation".
FROM golang:1.26-alpine

ARG GNMI_VERSION=v0.14.1

RUN apk add --no-cache openssl \
    && CGO_ENABLED=0 go install github.com/openconfig/gnmi/testing/fake/gnmi/cmd/fake_server@${GNMI_VERSION}

RUN mkdir -p /etc/gnmi \
    && openssl req -x509 -newkey rsa:2048 -nodes \
      -keyout /etc/gnmi/server.key -out /etc/gnmi/server.crt -days 36500 \
      -subj "/CN=gnmi-fake-target" \
      -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

COPY config.pb.txt /etc/gnmi/config.pb.txt

EXPOSE 9339

ENTRYPOINT ["/go/bin/fake_server", \
  "-config", "/etc/gnmi/config.pb.txt", \
  "-text", \
  "-port", "9339", \
  "-server_crt", "/etc/gnmi/server.crt", \
  "-server_key", "/etc/gnmi/server.key", \
  "-allow_no_client_auth", \
  "-logtostderr"]
