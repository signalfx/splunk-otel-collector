FROM golang:1.26.5-bookworm

WORKDIR /go/src/app

COPY go.mod .
COPY main.go .

RUN go get

RUN go build

USER 65532:65532

CMD /go/src/app/logging
