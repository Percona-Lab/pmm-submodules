FROM centos:7
RUN yum -y install iproute nc \
    && rm -rf /var/cache/yum

COPY pmm2-client.tar.gz /tmp/pmm2-client.tar.gz
RUN tar -zxpf /tmp/pmm2-client.tar.gz -C /tmp \
    && cd /tmp/pmm2-client-* \
    && cp bin/* /usr/bin/
    && ln -s /usr/bin/pid-watchdog /usr/bin/service \
    && rm -rf /tmp/pmm2-client-*

COPY entrypoint.sh /entrypoint.sh

USER nobody

WORKDIR /usr/local/percona/

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/pid-watchdog"]
