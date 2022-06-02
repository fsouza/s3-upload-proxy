FROM golang:1.18.3-alpine AS build

ENV  CGO_ENABLED 0
WORKDIR /code
ADD  . ./
RUN  go test ./...
RUN  go install

FROM alpine:3.16.0
RUN apk add --no-cache ca-certificates mailcap
COPY --from=build /go/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
