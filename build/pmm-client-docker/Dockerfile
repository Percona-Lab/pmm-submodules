FROM centos:7
RUN yum -y install iproute nc \
    && rm -rf /var/cache/yum

RUN curl -o /tmp/pmm-client.tar.gz https://www.percona.com/downloads/pmm/1.17.1/binary/tarball/pmm-client-1.17.1.tar.gz
RUN tar -zxpf /tmp/pmm-client.tar.gz -C /tmp \
    && cd /tmp/pmm-client-* \
    && bash ./install \
    && rm -rf /tmp/pmm-client-*

COPY pmm2-client.tar.gz /tmp/pmm2-client.tar.gz
RUN tar -zxpf /tmp/pmm2-client.tar.gz -C /tmp \
    && cd /tmp/pmm2-client-* \
    && rm -rf /tmp/pmm2-client-*

RUN install -d -o nobody -g nobody -m 0775 /usr/local/percona \
    touch /usr/local/percona/pmm2/config/pmm-agent.yaml \
    && chmod -R g+rwx /usr/local/percona \
    && chown -R nobody:nobody /usr/local/percona \
    && install -d -o nobody -g nobody -m 0775 /run /var/log /etc/rc.d/init.d

COPY entrypoint.sh /entrypoint.sh

USER nobody
WORKDIR /usr/local/percona/
ENTRYPOINT ["/entrypoint.sh"]
