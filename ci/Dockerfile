FROM alpine:3.13.5

RUN apk add --no-cache ca-certificates mailcap
ADD s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
