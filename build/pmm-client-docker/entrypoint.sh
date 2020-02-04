#!/bin/bash

set -o errexit
set -o xtrace

urlencode() {
    old_lc_collate=$LC_COLLATE
    LC_COLLATE=C

    local length="${#1}"
    for (( i = 0; i < length; i++ )); do
        local c="${1:i:1}"
        case $c in
            [a-zA-Z0-9.~_-]) printf "$c" ;;
            *) printf '%%%02X' "'$c" ;;
        esac
    done

    LC_COLLATE=$old_lc_collate
}

puzzle() {
    local args="${1}"
    shift
    for i in "${@}"
    do
        args=${args/$i/xxxxx}
    done
    echo ${args}
}

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

pmm2_start() {
    AGENT_CONFIG_FILE=/usr/local/percona/pmm2/config/pmm-agent.yaml
    rm -f ${AGENT_CONFIG_FILE}

    { set +x; } 2> /dev/null
    if [ -n "${PMM_USER}" ]; then
        echo '+ ARGS+=" --server-username=PMM_USER'
        ARGS+=" --server-username=${PMM_USER}"
    fi

    ARGS_TO_PUZZLE=( ${PMM_USER} ${PMM_PASSWORD} )

    echo '+ pmm-agent setup \
        --force \
        --config-file='"'${AGENT_CONFIG_FILE}'"' \
        --server-address='"'${PMM_SERVER}'"' \
        --server-insecure-tls \
        --container-id='"'${HOSTNAME}'"' \
        --container-name='"'${CLIENT_NAME}'"' \
        --ports-min='"'${CLIENT_PORT_MIN:-30100}'"' \
        --ports-max='"'${CLIENT_PORT_MAX:-30200}'"' \
        --listen-port='"'${CLIENT_PORT_LISTEN:-7777}'"' \
        '"'$(puzzle "${ARGS}" "${ARGS_TO_PUZZLE[@]}")'"' \
        '"'${CLIENT_ADDR:-$SRC_ADDR}'"' container '"'${CLIENT_NAME}'"''

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
        "${CLIENT_ADDR:-$SRC_ADDR}" container "${CLIENT_NAME}"

    if [ -n "${DB_USER}" ]; then
        echo '+ DB_ARGS+=" --username=DB_USER'
        ARGS_TO_PUZZLE+=("${DB_USER}")
        DB_ARGS+=" --username=${DB_USER}"
    fi
    if [ -n "${DB_PASSWORD}" ]; then
        echo '+ DB_ARGS+=" --password=DB_PASSWORD'
        ARGS_TO_PUZZLE+=("${DB_PASSWORD}")
        DB_ARGS+=" --password=${DB_PASSWORD}"
    fi
    set -x
    if [ -n "${DB_CLUSTER}" ]; then
        DB_ARGS+=" --cluster=${DB_CLUSTER}"
    fi

    if [ -n "${DB_HOST}" -a "${DB_PORT}" ]; then
        wait_for_port "${DB_HOST}" "${DB_PORT}"
    fi

    pmm-agent --config-file="${AGENT_CONFIG_FILE}" \
        --ports-min="${CLIENT_PORT_MIN:-30100}" \
        --ports-max="${CLIENT_PORT_MAX:-30200}" > /usr/local/percona/pmm2/pmm-agent-tmp.log 2>&1 &

    wait_for_url "http://127.0.0.1:7777"

    cat /usr/local/percona/pmm2/pmm-agent-tmp.log

    if [ -n "${DB_TYPE}" ]; then
        { set +x; } 2>/dev/null
        case "${DB_TYPE}" in
            mongodb )
                if [[ "${DB_ARGS}" =~ "--uri" ]]; then
                    DB_ARGS=$(echo ${DB_ARGS} | cut -d ' ' -f 2-)
                fi
                echo '+ pmm-admin add "'"${DB_TYPE}"'" \
                    --skip-connection-check \
                    --server-url="https://xxxxx:xxxxx@${PMM_SERVER}/" \
                    --server-insecure-tls \
                    '"'$(puzzle "${DB_ARGS}" "${ARGS_TO_PUZZLE[@]}") \
                    ${CLIENT_NAME} \
                    ${DB_HOST}:${DB_PORT}'"''

                pmm-admin add "${DB_TYPE}" \
                    --skip-connection-check \
                    --server-url=https://${PMM_USER}:${ENCODED_PASSWORD}@${PMM_SERVER}/ \
                    --server-insecure-tls \
                    ${DB_ARGS} \
                    "${CLIENT_NAME}" \
                    "${DB_HOST}:${DB_PORT}"
                ;;
            * )
                echo "+ pmm-admin add "${DB_TYPE}" \
                    --skip-connection-check \
                    --server-url="https://xxxxx:xxxxx@${PMM_SERVER}/" \
                    --server-insecure-tls \
                    $(puzzle "${DB_ARGS}" "${ARGS_TO_PUZZLE[@]}") \
                    ${CLIENT_NAME} \
                    ${DB_HOST}:${DB_PORT}"

                pmm-admin add "${DB_TYPE}" \
                    --skip-connection-check \
                    --server-url="https://${PMM_USER}:${ENCODED_PASSWORD}@${PMM_SERVER}/" \
                    --server-insecure-tls \
                    ${DB_ARGS} \
                    "${CLIENT_NAME}" \
                    "${DB_HOST}:${DB_PORT}"
                ;;
        esac
        set -x
        
    fi

    kill %1
    exec pmm-agent \
        --config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml \
        --ports-min="${CLIENT_PORT_MIN:-30100}" \
        --ports-max="${CLIENT_PORT_MAX:-30200}"
}

pmm_start() {
    cd /usr/local/percona/pmm-client
    { set +x; } 2>/dev/null
    if [ -n "${PMM_USER}" ]; then
        echo '+ ARGS+=" --server-user=PMM_USER"'
        ARGS+=" --server-user=${PMM_USER}"
    fi

    ARGS_TO_PUZZLE=( ${PMM_USER} ${PMM_PASSWORD} )
    
    echo '+ pmm-admin config \
        --skip-root \
        --force \
        --server "'"${PMM_SERVER}"'" \
        --server-insecure-ssl \
        --bind-address "'"${SRC_ADDR}"'" \
        --client-address "'"${SRC_ADDR}"'" \
        --client-name "'"${CLIENT_NAME}"'" \
        "'"$(puzzle "${ARGS}" "${ARGS_TO_PUZZLE[@]}")"'"'

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
        echo '+ DB_ARGS+=" --user DB_USER"'
        DB_ARGS+=" --user ${DB_USER}"
        ARGS_TO_PUZZLE+=("${DB_USER}")
    fi
    if [ -n "${DB_PASSWORD}" ]; then
        echo '+ DB_ARGS+=" --password DB_PASSWORD"'
        DB_ARGS+=" --password ${DB_PASSWORD}"
        ARGS_TO_PUZZLE+=("${DB_PASSWORD}")
    fi
    set -x
    if [ -n "${DB_PORT}" ]; then
        DB_ARGS+=" --port ${DB_PORT}"
    fi

    if [ -n "${DB_HOST}" -a "${DB_PORT}" ]; then
        wait_for_port "${DB_HOST}" "${DB_PORT}"
    fi

    if [ -n "${DB_TYPE}" ]; then
        { set +x; } 2>/dev/null
        case "${DB_TYPE}" in
            proxysql )
                echo '+ pmm-admin add "'"${DB_TYPE}"'" \
                            --skip-root \
                            --dsn "'"DB_USER:DB_PASSWORD@tcp(${DB_HOST}:${DB_PORT})/"'"'
                pmm-admin add "${DB_TYPE}" \
                    --skip-root \
                    --dsn "${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/" 
                ;;
            mongodb )
                if [[ "${DB_ARGS}" =~ "--uri" ]]; then
                    echo 'pmm-admin add "${DB_TYPE}" \
                        --skip-root \
                        "'"$(puzzle "${DB_ARGS}" "${ARGS_TO_PUZZLE[@]}")"'"'
                    pmm-admin add "${DB_TYPE}" \
                        --skip-root \
                        ${DB_ARGS}
                else
                    echo '+ pmm-admin add "'"${DB_TYPE}"'" \
                                --skip-root \
                                --uri "'"mongodb://DB_USER:DB_PASSWORD@${DB_HOST}:${DB_PORT}/"'"'
                    pmm-admin add "${DB_TYPE}" \
                        --skip-root \
                        --uri "mongodb://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/" 
                fi
                ;;
            * )
                : echo 'pmm-admin add "${DB_TYPE}" \
                    --skip-root \
                    "'"$(puzzle "${DB_ARGS}" "${ARGS_TO_PUZZLE[@]}")"'"'
                pmm-admin add "${DB_TYPE}" \
                    --skip-root \
                    ${DB_ARGS}
                ;;
        esac
        set -x
    fi
    exec pid-watchdog
}

main() {
    if [ -z "${PMM_SERVER}" ]; then
        echo PMM_SERVER is not specified. exiting
        exit 1
    fi
    { set +x; } 2> /dev/null
    if [ -n "${PMM_PASSWORD}" ]; then
        echo '+ ARGS+="--server-password=PMM_PASSWORD"'
        ARGS+="--server-password=${PMM_PASSWORD}"
    fi
    ENCODED_PASSWORD=$(urlencode ${PMM_PASSWORD})
    set -x


    PMM_SERVER_IP=$(ping -c 1 "${PMM_SERVER/:*/}" | grep PING | sed -e 's/).*//; s/.*(//')
    SRC_ADDR=$(ip route get "${PMM_SERVER_IP}" | grep 'src ' | sed -e 's/.* src //; s/ .*//')
    CLIENT_NAME="${CLIENT_NAME:-$HOSTNAME}"

    { set +x; } 2> /dev/null
    SERVER_RESPONSE_CODE=$(curl -k -s -o /dev/null -w "%{http_code}" "https://${PMM_USER}:$(urlencode ${PMM_PASSWORD})@${PMM_SERVER}/v1/readyz")
    set -x

    if [[ "${SERVER_RESPONSE_CODE}" == '200' ]]; then
        export PATH="/usr/local/percona/pmm2/bin/:${PATH}"
        pmm2_start
    else
        export PATH="/usr/local/percona/pmm-client/:/usr/local/percona/qan-agent/bin/:${PATH}"
        pmm_start
    fi

}

main
