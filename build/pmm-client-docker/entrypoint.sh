#!/bin/bash

set -o errexit
set -o xtrace

wait_for_url() {
    local URL=$1
    local RESPONSE=$2

    for i in `seq 1 60` ; do
        curl -k "${URL}" | grep "$RESPONSE" > /dev/null 2>&1 && result=$? || result=$?
        if [ $result -eq 0 ] ; then
            return
        fi
        sleep 1
    done

    echo "Operation timed out" >&2
    exit 1
}

wait_for_port() {
    local HOST=$1
    local PORT=$2

    for i in `seq 1 60` ; do
        nc -z "${HOST}" "${PORT}" > /dev/null 2>&1 && result=$? || result=$?
        if [ $result -eq 0 ] ; then
            return
        fi
        sleep 1
    done

    echo "Operation timed out" >&2
    exit 1
}

if [ -z "${PMM_SERVER}" ]; then
    echo PMM_SERVER is not specified. exiting
    exit 1
fi
if [ -n "${PMM_USER}" ]; then
    ARGS+=" --server-username=${PMM_USER}"
fi
if [ -n "${PMM_PASSWORD}" ]; then
    ARGS+=" --server-password=${PMM_PASSWORD}"
fi

PMM_SERVER_IP=$(ping -c 1 "${PMM_SERVER/:*/}" | grep PING | sed -e 's/).*//; s/.*(//')
SRC_ADDR=$(ip route get "${PMM_SERVER_IP}" | grep 'src ' | sed -e 's/.* src //; s/ .*//')
CLIENT_NAME=${DB_HOST:-$HOSTNAME}

wait_for_url "https://${PMM_USER}:${PMM_PASSWORD}@${PMM_SERVER}/v1/readyz"

pmm-agent setup \
  --force \
  --config-file=pmm-agent.yaml \
  --server-address=${PMM_SERVER} \
  --server-insecure-tls \
  --ports-min=41000 \
  --ports-max=41050 \
  ${ARGS} \
  ${CLIENT_ADDR:-$SRC_ADDR} container ${CLIENT_NAME}

exec "$@"
