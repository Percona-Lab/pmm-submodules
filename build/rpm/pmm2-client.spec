%define debug_package %{nil}
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
%if 0%{?rhel} > 6
Requires(post):   systemd
Requires(preun):  systemd
Requires(postun): systemd
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
%if 0%{?rhel} >= 7
install -m 755 -d $RPM_BUILD_ROOT/%{_unitdir}
install -m 644 config/pmm-agent.service %{buildroot}/%{_unitdir}/pmm-agent.service
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

%post
%if 0%{?rhel} >= 7
%systemd_post pmm-agent.service
  if [ $1 == 1 ]; then
    /usr/bin/systemctl enable pmm-agent >/dev/null 2>&1 || :
  fi
for file in node_exporter mysqld_exporter postgres_exporter mongodb_exporter proxysql_exporter
do
  %{__ln_s} -f /usr/local/percona/$file /usr/bin/$file
done
%endif

%preun
%if 0%{?rhel} >= 7
%systemd_preun pmm-agent.service
%endif

%postun
%if 0%{?rhel} >= 7
%systemd_postun pmm-agent.service
%endif
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
%{_unitdir}/pmm-agent.service
%endif
/usr/sbin/pmm-agent
/usr/local/percona/node_exporter
/usr/local/percona/mysqld_exporter
/usr/local/percona/postgres_exporter
/usr/local/percona/proxysql_exporter
/usr/local/percona/mongodb_exporter
%config(noreplace) /usr/local/percona/pmm-agent.yaml
