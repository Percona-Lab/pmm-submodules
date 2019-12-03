FROM registry.access.redhat.com/ubi7/ubi

RUN yum -y install iproute nc \
    && rm -rf /var/cache/yum

RUN curl -Lo /tmp/pmm-client.tar.gz https://www.percona.com/downloads/pmm/1.17.1/binary/tarball/pmm-client-1.17.1.tar.gz \
    && tar -zxpf /tmp/pmm-client.tar.gz -C /tmp \
    && cd /tmp/pmm-client-* \
    && bash ./install \
    && touch /usr/local/percona/pmm-client/config.yaml \
    && rm -rf /tmp/pmm-client-*

RUN curl -Lo /usr/bin/pid-watchdog https://github.com/percona/pid-watchdog/releases/download/v0.1.0/pid-watchdog \
    && chown nobody:root /usr/bin/pid-watchdog \
    && chmod u+rwx,g+rwx /usr/bin/pid-watchdog \
    && ln -s /usr/bin/pid-watchdog /usr/bin/service

RUN curl -Lo /tmp/pmm2-client.tar.gz https://www.percona.com/downloads/pmm2/2.1.0/binary/tarball/pmm2-client-2.1.0.tar.gz \
    && tar -zxpf /tmp/pmm2-client.tar.gz -C /tmp \
    && cd /tmp/pmm2-client-* \
    && bash -x ./install_tarball \
    && rm -rf /tmp/pmm2-client-*

RUN install -d -o nobody -g nobody -m 0775 /usr/local/percona \
    && chmod -R g+rwx /usr/local/percona \
    && chown -R nobody:root /usr/local/percona \
    && install -d -o nobody -g root -m 0775 /run /var/log /etc/rc.d/init.d

COPY entrypoint.sh /entrypoint.sh

USER nobody

EXPOSE 7777
EXPOSE 30100-30200

ENTRYPOINT ["/entrypoint.sh"]
