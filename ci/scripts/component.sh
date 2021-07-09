#!/bin/bash -eux

pushd dp-import-cantabular-dimension-options/features/compose
  docker compose up --abort-on-container-exit
popd
