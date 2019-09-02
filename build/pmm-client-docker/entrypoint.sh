#!/bin/bash

set -o errexit
set -o xtrace

AGENT_CONFIG_FILE=/usr/local/percona/pmm-agent.yaml

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
CLIENT_NAME=${CONTAINER_NAME:-$HOSTNAME}

wait_for_url "https://${PMM_USER}:${PMM_PASSWORD}@${PMM_SERVER}/v1/readyz"

rm -f /usr/local/percona/pmm-agent.yaml

pmm-agent setup \
  --force \
  --config-file=${AGENT_CONFIG_FILE} \
  --server-address=${PMM_SERVER} \
  --server-insecure-tls \
  --container-id=${HOSTNAME} \
  --container-name=${CLIENT_NAME} \
  --ports-min=${CLIENT_PORT_MIN:-30100} \
  --ports-max=${CLIENT_PORT_MAX:-30200} \
  --listen-port=${CLIENT_PORT_LISTEN:-7777} \
  ${ARGS} \
  ${CLIENT_ADDR:-$SRC_ADDR} container ${CLIENT_NAME}

if [ -n "${DB_USER}" ]; then
    DB_ARGS+=" --username=${DB_USER}"
fi
if [ -n "${DB_PASSWORD}" ]; then
    DB_ARGS+=" --password=${DB_PASSWORD}"
fi
if [ -n "${DB_CLUSTER}" ]; then
    DB_ARGS+=" --cluster=${DB_CLUSTER}"
fi

if [ -n "${DB_HOST}" -a "${DB_PORT}" ]; then
    wait_for_port "${DB_HOST}" "${DB_PORT}"
fi

cat ${AGENT_CONFIG_FILE}

pmm-agent --config-file=${AGENT_CONFIG_FILE} \
 --ports-min=${CLIENT_PORT_MIN:-30100} \
 --ports-max=${CLIENT_PORT_MAX:-30200} > /usr/local/percona/pmm-agent-tmp-log.log 2>&1 &

wait_for_url "http://127.0.0.1:7777/Status"

cat /usr/local/percona/pmm-agent-tmp-log.log

if [ -n "${DB_TYPE}" ]; then
    pmm-admin add "${DB_TYPE}" \
        --skip-connection-check \
        --server-url="https://${PMM_USER}:${PMM_PASSWORD}@${PMM_SERVER}/" \
        --server-insecure-tls \
        ${DB_ARGS} \
        ${DB_HOST}:${DB_PORT}
fi

kill %1

exec "$@"
