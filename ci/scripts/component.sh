#!/bin/bash -eux

pushd dp-import-cantabular-dimension-options/features/compose
  echo "hello component tests"
  docker-compose up --abort-on-container-exit
popd
