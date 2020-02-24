FROM golang:1

WORKDIR /go/src/app
ADD . /go/src/app
ENV GOOS=linux
RUN go test ./...
RUN go build -o /go/bin/app
