FROM golang:1.16.3-stretch

WORKDIR /go/src/app

COPY go.mod .
COPY main.go .

RUN go get

RUN go build

CMD /go/src/app/app
