FROM centos:7
WORKDIR /root/rpmbuild/SOURCES
COPY golang/* ./

RUN yum install -y \
                   yum-utils rpm-build rpmdevtools \
                   golang pcre-devel glibc-static
RUN spectool -g  go-srpm-macros.spec
RUN rpmbuild -ba go-srpm-macros.spec
RUN spectool -g  golang.spec
RUN rpmbuild -ba golang.spec


FROM centos:7
RUN yum install -y gcc gcc-c++ \
                   libtool libtool-ltdl \
                   make cmake \
                   git \
                   pkgconfig \
                   sudo \
                   automake autoconf \
                   rpmdevtools createrepo_c epel-release \
                   yum-utils rpm-build \
                   wget \
                   epel-release
RUN yum install -y nodejs

COPY --from=0 /root/rpmbuild/RPMS /var/cache/yum/RPMS
RUN yum install -y \
    /var/cache/yum/RPMS/*/golang-[123].*.rpm \
    /var/cache/yum/RPMS/*/golang-bin-[123].*.rpm \
    /var/cache/yum/RPMS/*/golang-src-[123].*.rpm \
    /var/cache/yum/RPMS/*/go-srpm-macros-*.rpm
RUN yum clean all && rm -rf /var/cache/yum

RUN useradd builder -u 1000 -m -G users,wheel && \
    echo "builder ALL=(ALL:ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    echo "# macros"                      >  /home/builder/.rpmmacros && \
    echo "%_topdir    /home/builder/rpm" >> /home/builder/.rpmmacros && \
    mkdir /home/builder/rpm && \
    chmod 755 /home/builder && \
    chown -R builder:builder /home/builder
USER builder

ENV FLAVOR=rpmbuild OS=centos DIST=el7
WORKDIR /home/builder/rpm
