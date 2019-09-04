%define debug_package %{nil}

%{!?with_systemd:%global systemd 0}
%{?el7:          %global systemd 1}
%{?el8:          %global systemd 1}

Name:           pmm2-client
Summary:        Percona Monitoring and Management Client
Version:        %{version}
Release:        %{release}%{?dist}
Group:          Applications/Databases
License:        ASL 2.0
Vendor:         Percona LLC
URL:            https://percona.com
Source:         pmm2-client-%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root
Summary:        PMM-agent

%if 0%{?systemd}
BuildRequires:  systemd
BuildRequires:  pkgconfig(systemd)
Requires(post):   systemd
Requires(preun):  systemd
Requires(postun): systemd
%else
Requires(post):   /sbin/chkconfig
Requires(preun):  /sbin/chkconfig
Requires(preun):  /sbin/service
%endif
AutoReq:        no
Conflicts:      pmm-client

%description
Percona Monitoring and Management (PMM) is an open-source platform for managing and monitoring MySQL and MongoDB
performance. It is developed by Percona in collaboration with experts in the field of managed database services,
support and consulting.
PMM is a free and open-source solution that you can run in your own environment for maximum security and reliability.
It provides thorough time-based analysis for MySQL and MongoDB servers to ensure that your data works as efficiently
as possible.


%prep
%setup -q


%build

%install
install -m 0755 -d $RPM_BUILD_ROOT/usr/sbin
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/bin
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/config
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/low-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/medium-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/high-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/low-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/medium-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/high-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/low-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/medium-resolution
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/high-resolution


install -m 0755 bin/pmm-admin $RPM_BUILD_ROOT/usr/local/percona/pmm2/bin
install -m 0755 bin/pmm-agent $RPM_BUILD_ROOT/usr/local/percona/pmm2/bin
install -m 0755 bin/node_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 bin/mysqld_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 bin/postgres_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 bin/mongodb_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 bin/proxysql_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0755 bin/rds_exporter $RPM_BUILD_ROOT/usr/local/percona/pmm2/exporters
install -m 0660 example.prom $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/low-resolution/
install -m 0660 example.prom $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/medium-resolution/
install -m 0660 example.prom $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/textfile-collector/high-resolution/
install -m 0660 queries-mysqld.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/low-resolution/
install -m 0660 queries-mysqld.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/medium-resolution/
install -m 0660 queries-mysqld.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/mysql/high-resolution/
install -m 0660 queries-postgres.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/low-resolution/
install -m 0660 queries-postgres.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/medium-resolution/
install -m 0660 queries-postgres.yml $RPM_BUILD_ROOT/usr/local/percona/pmm2/collectors/custom-queries/postgresql/high-resolution/
%if 0%{?systemd}
  install -m 0755 -d $RPM_BUILD_ROOT/%{_unitdir}
  install -m 0644 config/pmm-agent.service %{buildroot}/%{_unitdir}/pmm-agent.service
%else
  install -m 0755 -d $RPM_BUILD_ROOT/etc/rc.d/init.d
  install -m 0750 config/pmm-agent.init $RPM_BUILD_ROOT/etc/rc.d/init.d/pmm-agent
%endif



%clean
rm -rf $RPM_BUILD_ROOT

%pre
if [ $1 == 1 ]; then
  if ! getent passwd pmm-agent > /dev/null 2>&1; then
    /usr/sbin/groupadd -r pmm-agent
    /usr/sbin/useradd -M -r -g pmm-agent -d /usr/local/percona/ -s /bin/false -c pmm-agent pmm-agent > /dev/null 2>&1
  fi
fi
if [ $1 -eq 2 ]; then
    %if 0%{?systemd}
      /usr/bin/systemctl stop pmm-agent.service >/dev/null 2>&1 ||:
    %else
      /sbin/service pmm-agent stop >/dev/null 2>&1 ||:
    %endif
fi


%post
%if 0%{?systemd}
  %systemd_post pmm-agent.service
  if [ $1 == 1 ]; then
      if [ ! -f /usr/local/percona/pmm2/config/pmm-agent.yaml ]; then
          install -d -m 0755 /usr/local/percona/pmm2/config
          install -m 0640 -o pmm-agent -g pmm-agent /dev/null /usr/local/percona/pmm2/config/pmm-agent.yaml
      fi
      /usr/bin/systemctl enable pmm-agent >/dev/null 2>&1 || :
      /usr/bin/systemctl daemon-reload
      /usr/bin/systemctl start pmm-agent.service
  fi
%else
  if [ $1 == 1 ]; then
      install -m 0640 -o pmm-agent -g pmm-agent /dev/null /var/log/pmm-agent.log
      if [ ! -f /usr/local/percona/pmm2/config/pmm-agent.yaml ]; then
          install -d -m 0755 /usr/local/percona/pmm2/config
          install -m 0640 -o pmm-agent -g pmm-agent /dev/null /usr/local/percona/pmm2/config/pmm-agent.yaml
      fi
      /sbin/chkconfig --add pmm-agent
      /sbin/service pmm-agent start >/dev/null 2>&1 ||:
  fi
%endif

for file in pmm-admin pmm-agent
do
  %{__ln_s} -f /usr/local/percona/pmm2/bin/$file /usr/bin/$file
  %{__ln_s} -f /usr/local/percona/pmm2/bin/$file /usr/sbin/$file
done

if [ $1 -eq 2 ]; then
    %if 0%{?systemd}
      /usr/bin/systemctl daemon-reload
      /usr/bin/systemctl start pmm-agent.service
    %else
      /sbin/service pmm-agent start >/dev/null 2>&1 ||:
    %endif
fi

%preun
%if 0%{?rhel} >= 7
  %systemd_preun pmm-agent.service
%else
  if [ "$1" = 0 ]; then
    /sbin/service pmm-agent stop >/dev/null 2>&1 || :
    /sbin/chkconfig --del pmm-agent
  fi
%endif

%postun
case "$1" in
   0) # This is a yum remove.
      /usr/sbin/userdel pmm-agent
      %if 0%{?systemd}
          %systemd_postun_with_restart pmm-agent.service
      %endif
   ;;
   1) # This is a yum upgrade.
      %if 0%{?systemd}
          %systemd_postun_with_restart pmm-agent.service
      %else
          /sbin/service pmm-agent restart >/dev/null 2>&1 || :
      %endif
   ;;
esac
if [ $1 == 0 ]; then
  if /usr/bin/id -g pmm-agent > /dev/null 2>&1; then
    /usr/sbin/userdel pmm-agent > /dev/null 2>&1
    /usr/sbin/groupdel pmm-agent > /dev/null 2>&1 || true
    if [ -f /usr/local/percona/pmm2/config/pmm-agent.yaml ]; then
        rm -r /usr/local/percona/pmm2/config/pmm-agent.yaml
    fi
    for file in pmm-admin pmm-agent
    do
      if [ -L /usr/sbin/$file ]; then
        rm -rf /usr/sbin/$file
      fi
      if [ -L /usr/bin/$file ]; then
        rm -rf /usr/bin/$file
      fi
    done
  fi
fi


%files
%if 0%{?rhel} >= 7
%config %{_unitdir}/pmm-agent.service
%else
/etc/rc.d/init.d/pmm-agent
%endif
%attr(0660,pmm-agent,pmm-agent) %ghost /usr/local/percona/pmm2/config/pmm-agent.yaml
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/pmm2

%changelog
* Thu Aug 29 2019 Evgeniy Patlan <evgeniy.patlan@percona.com>
- Rework file structure.
