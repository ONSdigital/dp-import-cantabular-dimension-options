#!/bin/bash -eux

# Run component tests in docker compose
pushd dp-import-cantabular-dimension-options/features/compose
  docker-compose up --abort-on-container-exit
  e=$?
popd

# Show the output, which was stored in $COMPONENT_TEST_LOG_FILE
f=${COMPONENT_TEST_LOG_FILE-""}
if [ -n "$f" ]; then
  cat $f && rm $f
fi
echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"

# exit with the same code returned by docker compose
exit $e
