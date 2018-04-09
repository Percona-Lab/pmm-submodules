%global commit        e2fc0f50fafe52ff3cbc7820f13c16a3b5d4af0d
%global shortcommit   %(c=%{commit}; echo ${c:0:7})

Name:           go-srpm-macros
Version:        2
Release:        16%{?dist}
Summary:        RPM macros for building Golang packages for various architectures
Group:          Development/Libraries
License:        GPLv3+
Source0:        https://github.com/gofed/go-macros/archive/%{commit}/go-macros-%{shortcommit}.tar.gz
BuildArch:      noarch
# for install command
BuildRequires:  coreutils

%description
The package provides macros for building projects in Go
on various architectures.

%prep
%setup -q -n go-macros-%{commit}

%build
# nothing to build, just for hooks

%install
install -m 644 -D rpm/macros.d/macros.go-srpm %{buildroot}%{_rpmconfigdir}/macros.d/macros.go-srpm
%if 0%{?fedora} < 29
# Use macros.forge222 so it does not conflict with macros.forge from the redhat-rpm-config
install -m 644 -D rpm/macros.d/macros.forge %{buildroot}%{_rpmconfigdir}/macros.d/macros.forge222
%endif

%files
%{_rpmconfigdir}/macros.d/macros.go-srpm
%if 0%{?fedora} < 29
%{_rpmconfigdir}/macros.d/macros.forge222
%endif

%changelog
* Mon Mar 05 2018 Jan Chaloupka <jchaloup@redhat.com> - 2-16
- Switch to upstream tarball (2nd attempt)

* Sun Mar 04 2018 Jan Chaloupka <jchaloup@redhat.com> - 2-15
- Build the rawhide gometa completely on rawhide forgemeta

* Tue Feb 27 2018 Robert-André Mauchin <zebob.m@gmail.com> - 2-14
- Fix the Github download path

* Fri Feb 23 2018 Jan Chaloupka <jchaloup@redhat.com> - 2-13
- Update only the macros.go-srpm file, the upstream tarball can not be found

* Fri Feb 23 2018 Jan Chaloupka <jchaloup@redhat.com> - 2-12
- Install go-srpm macros from an upstream tarball

* Wed Feb 07 2018 Fedora Release Engineering <releng@fedoraproject.org> - 2-11
- Rebuilt for https://fedoraproject.org/wiki/Fedora_28_Mass_Rebuild

* Wed Jul 26 2017 Fedora Release Engineering <releng@fedoraproject.org> - 2-10
- Rebuilt for https://fedoraproject.org/wiki/Fedora_27_Mass_Rebuild

* Wed Jul 12 2017 Jakub Čajka <jcajka@redhat.com> - 2-9
- Drop ppc64 from go arches
- https://fedoraproject.org/wiki/Changes/golang1.9

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
