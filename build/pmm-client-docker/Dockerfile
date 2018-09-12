FROM centos:latest

ENV MONIT_VERSION=5.25.2

# Compile and install monit
RUN yum -y install make gcc curl zlib-devel openssl-devel file \
    && cd /tmp \
    && curl -sS https://mmonit.com/monit/dist/monit-${MONIT_VERSION}.tar.gz | gunzip -c - | tar -xf - \
    && cd monit-${MONIT_VERSION} \
    && ./configure --without-pam \
    && make \
    && make install

COPY pmm-client.tar.gz /tmp/pmm-client.tar.gz
RUN tar -zxpf /tmp/pmm-client.tar.gz -C /tmp \
    && cd /tmp/pmm-client-* \
    && bash ./install


FROM centos:latest
RUN yum -y install iproute \
    && rm -rf /var/cache/yum

COPY --from=0 /usr/local/bin/monit /usr/bin/monit
ADD --chown=nobody:nobody monitrc /etc/monitrc
RUN chmod 700 /etc/monitrc
RUN install -d -o nobody -g nobody /etc/monit.d /var/lib/monit /var/lib/monit/eventqueue /run /var/log /etc/rc.d/init.d

COPY --from=0 --chown=nobody:nobody /usr/local/percona /usr/local/percona
COPY --from=0 --chown=nobody:nobody /usr/sbin/pmm-admin /usr/sbin/pmm-admin
ADD service /usr/bin/service
ADD entrypoint.sh /entrypoint.sh

USER nobody
ENTRYPOINT ["/entrypoint.sh"]
