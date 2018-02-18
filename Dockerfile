FROM golang:1.10-alpine AS build
ENV  CGO_ENABLED 0
ADD  . /go/src/github.com/fsouza/s3-upload-proxy
RUN  go test github.com/fsouza/s3-upload-proxy
RUN  go install github.com/fsouza/s3-upload-proxy

FROM alpine:3.7
RUN apk add --no-cache ca-certificates mailcap
COPY --from=build /go/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
