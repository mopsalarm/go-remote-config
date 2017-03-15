#!/bin/sh

set -e

CGO_ENABLED=0 ./build.sh

docker build -t mopsalarm/go-remote-config .
docker push mopsalarm/go-remote-config
