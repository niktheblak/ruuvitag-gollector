FROM golang:1.16

VOLUME /go/src/app
WORKDIR /go/src/app
ADD go.mod .
ADD go.sum .
RUN go mod download
ENV GOOS=linux
ENTRYPOINT ["go"]
