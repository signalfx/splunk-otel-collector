FROM golang:1.18.1-stretch

WORKDIR /go/src/app

COPY go.mod .
COPY main.go .

RUN go get

RUN go build

CMD /go/src/app/app
