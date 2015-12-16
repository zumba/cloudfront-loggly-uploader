#! /usr/bin/env bash
#
# Build Linux RPM and Deb packages using fpm
#

set -o nounset
set -o errexit

PKG_NAME="cloudfront-loggly-uploader"
MAINTAINER="Rene Fragoso"
VENDOR="Zumba"
URL="https://github.com/zumba/cloudfront-loggly-uploader"
VERSION="$(date +%s)"

PWD="$(pwd)"
BUILD="${PWD}/cloudfront-loggly-uploader-${VERSION}"
TARGET_ARCH="x86_64"

mkdir -p "${BUILD}/opt/cloudfront-loggly-uploader"
mkdir -p "${BUILD}/etc"

cp ../cloudfront-loggly-uploader "${BUILD}/opt/cloudfront-loggly-uploader"
cp ../sample.conf.yaml "${BUILD}/etc/cloudfront-loggly-uploader.conf.yaml"

# Build RPM
fpm -s dir \
  -t rpm \
  -v "${VERSION}" \
  -n "${PKG_NAME}" \
  --maintainer="${MAINTAINER}" \
  --vendor "${VENDOR}" \
  --url="${URL}" \
  --category Network \
  --description "Download AWS CloudFront logs from S3 and upload them to Loggly" \
  -a "${TARGET_ARCH}" \
  -C "${BUILD}" \
  --config-files "etc/cloudfront-loggly-uploader.conf.yaml" \
  ./
  
# Build Deb
fpm -s dir \
  -t deb \
  -v "${VERSION}" \
  -n "${PKG_NAME}" \
  --maintainer="${MAINTAINER}" \
  --vendor "${VENDOR}" \
  --url="${URL}" \
  --category Network \
  --description "Download AWS CloudFront logs from S3 and upload them to Loggly" \
  -a "${TARGET_ARCH}" \
  -C "${BUILD}" \
  --config-files "etc/cloudfront-loggly-uploader.conf.yaml" \
  ./

# Clean up
rm -rf "${BUILD}"
