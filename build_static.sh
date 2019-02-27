#!/bin/sh

docker run --rm -i -v `pwd`:/gopath/src/github.com/3ft9/qremlin stut/go-build-centos7.4:latest /bin/sh << 'EOF'
set -ex
cd /gopath/src/github.com/3ft9/qremlin
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/gopath
make build
strip qremlin
EOF
