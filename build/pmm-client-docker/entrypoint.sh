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
    ARGS+=" --server-user ${PMM_USER}"
fi
if [ -n "${PMM_PASSWORD}" ]; then
    ARGS+=" --server-password ${PMM_PASSWORD}"
fi

PMM_SERVER_IP=$(ping -c 1 "${PMM_SERVER/:*/}" | grep PING | sed -e 's/).*//; s/.*(//')
SRC_ADDR=$(ip route get "${PMM_SERVER_IP}" | grep 'src ' | sed -e 's/.* src //; s/ .*//')
CLIENT_NAME=${DB_HOST:-$HOSTNAME}

wait_for_url "https://${PMM_USER}:${PMM_PASSWORD}@${PMM_SERVER}/qan-api/ping" ""
wait_for_url "https://${PMM_USER}:${PMM_PASSWORD}@${PMM_SERVER}/v1/status/leader" "127.0.0.1:8300"

pmm-admin config \
    --skip-root \
    --force \
    --server "${PMM_SERVER}" \
    --server-insecure-ssl \
    --bind-address "${SRC_ADDR}" \
    --client-address "${SRC_ADDR}" \
    --client-name "${CLIENT_NAME}" \
    ${ARGS}

if [ -n "${DB_HOST}" ]; then
    DB_ARGS+=" --host ${DB_HOST}"
fi
if [ -n "${DB_USER}" ]; then
    DB_ARGS+=" --user ${DB_USER}"
fi
if [ -n "${DB_PASSWORD}" ]; then
    DB_ARGS+=" --password ${DB_PASSWORD}"
fi
if [ -n "${DB_PORT}" ]; then
    DB_ARGS+=" --port ${DB_PORT}"
fi

if [ -n "${DB_HOST}" -a "${DB_PORT}" ]; then
    wait_for_port "${DB_HOST}" "${DB_PORT}"
fi

if [ -n "${DB_TYPE}" ]; then
    pmm-admin add "${DB_TYPE}" \
        --skip-root \
        ${DB_ARGS}
fi

exec /usr/bin/monit -I
