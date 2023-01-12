FROM golang:1.19

WORKDIR /go/src/app
ADD go.mod .
ADD go.sum .
RUN go mod download
ENV GOOS=linux
ENTRYPOINT ["go"]
