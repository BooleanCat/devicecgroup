#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$( dirname "$DIR" )"

docker run \
  --privileged=true \
  -v $PROJECT_DIR:/root/go/src/github.com/BooleanCat/devicecgroup \
  -w /root/go/src/github.com/BooleanCat/devicecgroup \
  -i -t devicecgroup \
  ginkgo -r --race --randomizeAllSpecs
