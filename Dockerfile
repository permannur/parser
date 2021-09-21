# syntax=docker/dockerfile:1

FROM golang:1.17

WORKDIR /go/src/parser

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN go build

EXPOSE 8080

CMD [ "parser"]
