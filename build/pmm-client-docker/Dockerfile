FROM centos:7

RUN yum -y install iproute nc \
    && rm -rf /var/cache/yum

COPY pmm2-client.tar.gz /tmp/pmm2-client.tar.gz

RUN tar -zxpf /tmp/pmm2-client.tar.gz -C /tmp \
    && cd /tmp/pmm2-client-* \
    && cp bin/* /usr/bin/ \
    && rm -rf /tmp/pmm2-client-*

RUN install -d -o nobody -g nobody -m 0775 /usr/local/percona/pmm2/config \
    && chmod -R g+rwx /usr/local/percona \
    && chown -R nobody:nobody /usr/local/percona \
    && install -d -o nobody -g nobody -m 0775 /run /var/log /etc/rc.d/init.d \
    && install -d -o nobody -g nobody -m 0775 /usr/local/percona/pmm2/collectors/textfile-collector/high-resolution \
     /usr/local/percona/pmm2/collectors/textfile-collector/medium-resolution \
     /usr/local/percona/pmm2/collectors/textfile-collector/low-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/mysql/high-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/mysql/medium-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/mysql/low-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/postgres/high-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/postgres/medium-resolution \
     /usr/local/percona/pmm2/collectors/custom-queries/postgres/low-resolution

RUN install -d -o nobody -g nobody -m 0775 /usr/local/percona/pmm2/exporters && \
    cp /usr/bin/*_exporter /usr/local/percona/pmm2/exporters/

COPY entrypoint.sh /entrypoint.sh

USER nobody

WORKDIR /usr/local/percona/pmm2/

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 7777
EXPOSE 30100-30200

CMD ["pmm-agent", "--config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml", "--ports-min=30100", "--ports-max=30200"]
