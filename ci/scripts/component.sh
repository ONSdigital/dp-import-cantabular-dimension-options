#!/bin/bash -eux

pushd dp-import-cantabular-dimension-options/features/compose
  ls -la
  # Run component tests in docker compose
  docker-compose up --abort-on-container-exit
  e=$?
popd

ls -la

pushd dp-import-cantabular-dimension-options
  ls -la
  # Assuming that component-test output was forwareded to component-output.txt, display it
  cat component-output.txt && rm component-output.txt
popd

# Show message to prevent any confusion by 'ERROR 0' outpout
echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"

# exit with the same code returned by docker compose
exit $e
