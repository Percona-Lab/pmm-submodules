#!/bin/bash

set -o errexit
set -o xtrace

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

PMM_SERVER_IP=$(ping -c 1 "${PMM_SERVER}" | grep PING | sed -e 's/).*//; s/.*(//')
SRC_ADDR=$(ip route get "${PMM_SERVER_IP}" | grep 'src ' | awk '{print$7}')
CLIENT_NAME=${DB_HOST:-$HOSTNAME}

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

if [ -n "${DB_TYPE}" ]; then
    pmm-admin add "${DB_TYPE}" \
        --skip-root \
        ${DB_ARGS}
fi

exec /usr/bin/monit -I
