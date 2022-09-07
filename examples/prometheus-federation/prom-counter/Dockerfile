FROM golang:1.19.1-bullseye

WORKDIR /go/src/app

COPY go.mod .
COPY main.go .

RUN go get

RUN go build

CMD /go/src/app/app
