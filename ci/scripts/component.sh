#!/bin/bash -eux

# Run component tests in docker compose defined in features/compose folder
pushd dp-import-cantabular-dimension-options/features/compose
  docker-compose up --abort-on-container-exit
  e=$?
popd

# Cat the component-test output file and remove it, which is stored int he project folder
pushd dp-import-cantabular-dimension-options
  cat component-output.txt && rm component-output.txt
popd

# Show message to prevent any confusion by 'ERROR 0' outpout
echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"

# exit with the same code returned by docker compose
exit $e
