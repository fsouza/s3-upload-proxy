FROM golang:1.19.5-alpine AS build

ENV  CGO_ENABLED 0
WORKDIR /code
ADD  . ./
RUN  go test ./...
RUN  go install
RUN apk add --no-cache mailcap

FROM gcr.io/distroless/static
COPY --from=build /go/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
COPY --from=build /etc/mime.types /etc/mime.types
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
