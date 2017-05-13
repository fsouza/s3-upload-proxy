s3-upload-proxy
===============

Tool for proxying HTTP uploads to S3 buckets. Useful for private network
protected environments.

Running locally
---------------

Make sure you have [latest Go](https://golang.org/doc/install), then make sure
you have AWS credentials properly configured (s3-upload-proxy uses the [default
credential provider
chain](https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-default),
so you can use environment variables or file-based configuration).

Having Go and AWS credentials, just set the environment variable
``BUCKET_NAME``, build and start the process:

```
% export BUCKET_NAME=some-bucket
% go build -o s3-upload-proxy
% ./s3-upload-proxy
```

Environment variables
---------------------

s3-upload-proxy configuration's is defined using the following environment
variables:

| Variable         | Default value | Required  |
| ---------------- | ------------- | --------- |
| BUCKET_NAME      |               | Yes       |
| HEALTHCHECK_PATH | /healthcheck  | No        |
| HTTP_PORT        | 80            | No        |
| LOG_LEVEL        | debug         | No        |

Also available on Docker Hub: https://hub.docker.com/r/fsouza/s3-upload-proxy/.
