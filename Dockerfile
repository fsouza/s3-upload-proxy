FROM golang:1.11.3-alpine AS build
RUN apk add --no-cache git
ENV  CGO_ENABLED 0
ADD  . /code
WORKDIR /code
RUN  go install

FROM alpine:3.8
RUN apk add --no-cache ca-certificates mailcap
COPY --from=build /go/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
