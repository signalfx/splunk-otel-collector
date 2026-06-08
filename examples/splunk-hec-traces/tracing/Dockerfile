FROM golang:1.26.4-bookworm

WORKDIR /go/src/app

COPY go.mod .
COPY main.go .

RUN go get

RUN go build

CMD /go/src/app/tracing
