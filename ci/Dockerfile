FROM alpine:3.12.0 AS compressor

RUN apk add --no-cache upx
ADD s3-upload-proxy /usr/bin/s3-upload-proxy
RUN upx --brute /usr/bin/s3-upload-proxy

FROM alpine:3.12.0

RUN apk add --no-cache ca-certificates mailcap
COPY --from=compressor /usr/bin/s3-upload-proxy /usr/bin/s3-upload-proxy
ENTRYPOINT ["/usr/bin/s3-upload-proxy"]
