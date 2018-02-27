Name:           go-srpm-macros
Version:        2
Release:        8%{?dist}
Summary:        RPM macros for building Golang packages for various architectures
Group:          Development/Libraries
License:        GPLv3+
Source0:        macros.go-srpm
BuildArch:      noarch
# for install command
BuildRequires:  coreutils

%description
The package provides macros for building projects in Go
on various architectures.

%prep
# nothing to prep, just for hooks

%build
# nothing to build, just for hooks

%install
install -m 644 -D "%{SOURCE0}" \
    '%{buildroot}%{_rpmconfigdir}/macros.d/macros.go-srpm'

%files
%{_rpmconfigdir}/macros.d/macros.go-srpm

%changelog
* Fri Feb 10 2017 Fedora Release Engineering <releng@fedoraproject.org> - 2-8
- Rebuilt for https://fedoraproject.org/wiki/Fedora_26_Mass_Rebuild

* Wed Jul 20 2016 Jakub Čajka <jcajka@redhat.com> - 2-7
- move s390x to golang
- Related: bz1357394

* Wed Feb 03 2016 Fedora Release Engineering <releng@fedoraproject.org> - 2-6
- Rebuilt for https://fedoraproject.org/wiki/Fedora_24_Mass_Rebuild

* Thu Jan 28 2016 Jakub Čajka <jcajka@redhat.com> - 2-5
- move {power64} to golang

* Wed Dec 30 2015 Michal Toman <mtoman@fedoraproject.org> - 2-4
- MIPS has gcc-go, mips macro since rpm-4.12.0.1-18
  resolves: #1294875

* Thu Sep 10 2015 jchaloup <jchaloup@redhat.com> - 2-3
- Remove compiler specific macros (moved to go-compiler package)
- Define go-compiler macro to signal go-compiler packages is available

* Sat Aug 29 2015 jchaloup <jchaloup@redhat.com> - 2-2
- Add -ldflags $LDFLAGS to go build/test macro

* Sun Aug 23 2015 Peter Robinson <pbrobinson@fedoraproject.org> 2-1
- aarch64 now has golang

* Tue Jul 07 2015 jchaloup <jchaloup@redhat.com> - 1-1
- Initial commit
  resolves: #1241156
