FROM golang:1

VOLUME /go/src/app
WORKDIR /go/src/app
ENV GOOS=linux
ENTRYPOINT ["go"]
