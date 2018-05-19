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

| Variable            | Default value | Required | Description                                                                             |
| ------------------- | ------------- | -------- | --------------------------------------------------------------------------------------- |
| BUCKET_NAME         |               | Yes      | Name of the bucket                                                                      |
| HEALTHCHECK_PATH    | /healthcheck  | No       | Path for healthcheck                                                                    |
| HTTP_PORT           | 80            | No       | Port to bind (unsigned int)                                                             |
| LOG_LEVEL           | debug         | No       | Logging level                                                                           |
| CACHE_CONTROL_RULES |               | No       | JSON array with cache control rules (see below)                                         |

Defining cache-control rules
----------------------------

The tool also allow configuration for cache-control rules. The value of the
environment variable ``CACHE_CONTROL_RULES`` is a JSON array with the rules. An
example:

```
% export CACHE_CONTROL_RULES='[{"regexp":".mp4$","value":"public, max-age=3600"},{"regexp":".ts$","value":"public, max-age=2, s-maxage=999999"},{"regexp":".m3u8$","value":"private"}]'
```

Notice that the extension must include the dot.

Also available on Docker Hub: https://hub.docker.com/r/fsouza/s3-upload-proxy/.
