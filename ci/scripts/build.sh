#!/bin/bash -eux

pushd dp-import-cantabular-dimension-options
  make build
  cp build/dp-import-cantabular-dimension-options Dockerfile.concourse ../build
popd
