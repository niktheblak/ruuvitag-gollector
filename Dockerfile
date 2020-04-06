FROM golang:1

VOLUME /go/src/app
WORKDIR /go/src/app
ADD . .
RUN go get ./...
ENV GOOS=linux
ENTRYPOINT ["go"]
