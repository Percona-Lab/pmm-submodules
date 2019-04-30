%define debug_package %{nil}

%{!?with_systemd:%global systemd 0}
%{?el7:          %global systemd 1}
%{?el8:          %global systemd 1}

Name:           pmm2-client
Summary:        Percona Monitoring and Management Client
Version:        %{version}
Release:        %{release}%{?dist}
Group:          Applications/Databases
License:        AGPLv3
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
install -m 0755 bin/pmm-admin $RPM_BUILD_ROOT/usr/sbin/
install -m 0755 bin/pmm-agent $RPM_BUILD_ROOT/usr/sbin/
install -m 0755 -d $RPM_BUILD_ROOT/usr/local/percona/qan-agent/bin
install -m 0755 bin/node_exporter $RPM_BUILD_ROOT/usr/local/percona/
install -m 0755 bin/mysqld_exporter $RPM_BUILD_ROOT/usr/local/percona/
install -m 0755 bin/postgres_exporter $RPM_BUILD_ROOT/usr/local/percona/
install -m 0755 bin/mongodb_exporter $RPM_BUILD_ROOT/usr/local/percona/
install -m 0755 bin/proxysql_exporter $RPM_BUILD_ROOT/usr/local/percona/
install -m 0755 config/pmm-agent.yaml $RPM_BUILD_ROOT/usr/local/percona/
%if 0%{?systemd}
  install -m 755 -d $RPM_BUILD_ROOT/%{_unitdir}
  install -m 644 config/pmm-agent.service %{buildroot}/%{_unitdir}/pmm-agent.service
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
      /usr/bin/systemctl enable pmm-agent >/dev/null 2>&1 || :
      /usr/bin/systemctl daemon-reload
      /usr/bin/systemctl start pmm-agent.service
  fi
%else
  if [ $1 == 1 ]; then
      install -m 0640 -o pmm-agent -g pmm-agent /dev/null /var/log/pmm-agent.log
      /sbin/chkconfig --add pmm-agent
      /sbin/service pmm-agent start >/dev/null 2>&1 ||:
  fi
%endif

for file in node_exporter mysqld_exporter postgres_exporter mongodb_exporter proxysql_exporter
do
  %{__ln_s} -f /usr/local/percona/$file /usr/bin/$file
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
    for file in node_exporter mysqld_exporter postgres_exporter mongodb_exporter proxysql_exporter
    do
      if [ -L /usr/bin/$file ]; then
        rm -rf /usr/bin/$file
      fi
    done
  fi
fi


%files
/usr/sbin/pmm-admin
%if 0%{?rhel} >= 7
%config %{_unitdir}/pmm-agent.service
%else
/etc/rc.d/init.d/pmm-agent
%endif
%attr(-,pmm-agent,pmm-agent) /usr/sbin/pmm-agent
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/node_exporter
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/mysqld_exporter
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/postgres_exporter
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/proxysql_exporter
%attr(-,pmm-agent,pmm-agent) /usr/local/percona/mongodb_exporter
%attr(0660,pmm-agent,pmm-agent) %config(noreplace) /usr/local/percona/pmm-agent.yaml
