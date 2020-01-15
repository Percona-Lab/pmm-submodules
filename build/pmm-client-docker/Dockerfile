FROM registry.access.redhat.com/ubi7/ubi

COPY pmm2-client.tar.gz /tmp/pmm2-client.tar.gz

# need to use IDs for openshift
RUN groupadd -g 1002 pmm-agent
RUN useradd -u 1002 -r -g pmm-agent -s /sbin/nologin \
            -d /usr/local/percona/pmm2 \
            -c "PMM 2.X Client User" pmm-agent

RUN install -d -o pmm-agent -g pmm-agent -m 0775 /usr/local/percona/pmm2

RUN tar -zxpf /tmp/pmm2-client.tar.gz -C /tmp \
    && cd /tmp/pmm2-client-* \
    && env PMM_USER=pmm-agent PMM_GROUP=pmm-agent ./install_tarball \
    && cd /tmp \
    && rm -rf /tmp/pmm2-client-*

COPY entrypoint.py /entrypoint.py

USER pmm-agent
WORKDIR /usr/local/percona/pmm2/
ENTRYPOINT ["/entrypoint.py"]
