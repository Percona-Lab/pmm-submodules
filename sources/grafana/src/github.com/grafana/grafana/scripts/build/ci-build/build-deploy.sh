#!/bin/bash
set -eo pipefail

_version="1.2.28"
_tag="grafana/build-container:${_version}"

_dpath=$(dirname "${BASH_SOURCE[0]}")
cd "$_dpath"

docker build -t $_tag .
docker push $_tag
