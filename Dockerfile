FROM golang:1.14 as build-env
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...
RUN go build -v -o /go/bin/app

FROM gcr.io/distroless/static-debian10
COPY --from=build-env /go/bin/app /
CMD ["/app"]
