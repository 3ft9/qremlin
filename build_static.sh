#!/bin/sh

docker run -i -v `pwd`:/gopath/src/github.com/3ft9/qremlin alpine:edge /bin/sh << 'EOF'
set -ex
# Install prerequisites for the build process.
apk update
apk add git go libc-dev make
# Build qremlin.
cd /gopath/src/github.com/3ft9/qremlin
export GOPATH=/gopath
make build
strip qremlin
EOF
