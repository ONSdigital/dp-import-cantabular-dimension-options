#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-import-cantabular-dimension-options
  make audit
popd