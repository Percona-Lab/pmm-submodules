FROM centos:7

EXPOSE 80 443

WORKDIR /opt

RUN useradd -s /bin/false pmm
RUN yum -y install epel-release && yum -y install ansible

COPY RPMS /tmp/RPMS
COPY gitCommit /tmp/gitCommit
COPY playbook-install.yml /opt/playbook-install.yml
RUN ansible-playbook -vvv -i 'localhost,' -c local /opt/playbook-install.yml

COPY alertmanager.yml /etc/alertmanager.yml
COPY alertRules.yml /etc/alertRules.yml

COPY playbook-init.yml /opt/playbook-init.yml
RUN ansible-playbook -vvv -i 'localhost,' -c local /opt/playbook-init.yml
RUN cp /usr/share/pmm-server/entrypoint.sh /opt/entrypoint.sh

CMD ["/opt/entrypoint.sh"]
