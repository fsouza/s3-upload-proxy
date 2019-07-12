FROM golang:1.12.7-alpine AS build
RUN apk add --no-cache git
ENV  GOPROXY https://proxy.golang.org
ENV  CGO_ENABLED 0
ADD  . /code
WORKDIR /code
RUN  go test ./...
RUN  go install

FROM alpine:3.10.1
RUN apk add --no-cache ca-certificates mailcap
COPY --from=build /go/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
