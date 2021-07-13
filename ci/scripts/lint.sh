#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-import-cantabular-dimension-options
  make lint
popd
