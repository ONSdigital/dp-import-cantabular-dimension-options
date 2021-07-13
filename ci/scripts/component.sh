#!/bin/bash -eux

pushd dp-import-cantabular-dimension-options/features/compose
  docker-compose up --abort-on-container-exit
  echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"
popd