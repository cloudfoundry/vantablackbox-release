#!/bin/bash
set -e

RELEASE_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

pushd "${RELEASE_ROOT}"
  cp ../garden-ci/config/private.yml config/
  bosh upload-blobs
popd

