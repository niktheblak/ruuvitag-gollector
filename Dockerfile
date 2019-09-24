FROM golang:1.13 as build-env

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN go build -o /go/bin/app cmd/collector/main.go

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/app /
ADD configs/* .
CMD ["/app"]
