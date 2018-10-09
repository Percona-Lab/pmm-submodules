FROM centos:7
RUN yum -y install iproute nc \
    && rm -rf /var/cache/yum

COPY pmm-client.tar.gz /tmp/pmm-client.tar.gz
RUN tar -zxpf /tmp/pmm-client.tar.gz -C /tmp \
    && cd /tmp/pmm-client-* \
    && bash ./install \
    && cp bin/pid-watchdog /usr/bin/ \
    && ln -s /usr/bin/pid-watchdog /usr/bin/service \
    && rm -rf /tmp/pmm-client-*

RUN touch /usr/local/percona/pmm-client/config.yaml \
    && chmod -R g+rwx /usr/local/percona \
    && chown -R nobody:nobody /usr/local/percona \
    && install -d -o nobody -g nobody -m 0775 /run /var/log /etc/rc.d/init.d

COPY entrypoint.sh /entrypoint.sh

USER nobody
WORKDIR /usr/local/percona/pmm-client
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/pid-watchdog"]
