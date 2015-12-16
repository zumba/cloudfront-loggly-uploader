# cloudfront-loggly-uploader

This is a utility that downloads AWS CloudFront logs from S3 and uploads them to Loggly (written in Go). 

**Log files on S3 are deleted after successful upload to Loggly.**

## Building / Installing

```shell
go get github.com/zumba/cloudfront-loggly-uploader
cd $GOPATH/src/github.com/zumba/cloudfront-loggly-uploader
go build  # or go install
```

To build for another platform:
```shell
GOOS=linux go build
```

## Creating RPM or Debian packages

Creates an RPM and Debian package of the built binary. This uses [fpm](https://github.com/jordansissel/fpm) to create the packages.

```shell
go build
cd build/
./build-packages.sh
```

## Running

Utility was meant to be run in two ways: Either single use, or as a daemon.

For single use just run:

```shell
cloudfront-loggly-uploader -configFile cloudfront-loggly-uploader.conf.yaml
```

To run as a daemon, add the `-daemon` flag. It will continually check if new log files need to
be uploaded every X minutes (`daemon-sleep-mins` in config file). Add `-syslog` flag to log to
local syslog server instead of stdout.

```shell
cloudfront-loggly-uploader -configFile cloudfront-loggly-uploader.conf.yaml -daemon -syslog
```

## Sample config file

```yaml
daemon-sleep-mins: 5
s3-region: us-east-1
s3-bucket: cloudfront-logs
s3-prefix: cf-logs/
s3-aws-access-key-id: YOUR_AWS_ACCESS_KEY_ID
s3-aws-secret-access-key: YOUR_AWS_SECRET_ACCESS_KEY
loggly-api-key: YOUR_LOGGLY_API_KEY
loggly-tag: staging
```
