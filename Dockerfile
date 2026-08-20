FROM golang:1.27

ENV GOOS=linux

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

ENTRYPOINT ["go"]
