FROM centos:7
WORKDIR /root/rpmbuild/SOURCES
COPY golang/* ./

RUN yum install -y \
                   yum-utils rpm-build rpmdevtools \
                   golang pcre-devel glibc-static

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

RUN yum clean all && rm -rf /var/cache/yum

ENV GO_TARBALL https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz
RUN wget ${GO_TARBALL} -O /tmp/golang.tar.gz
RUN tar -C /usr/local -xzf /tmp/golang.tar.gz; \
    update-alternatives --install "/usr/bin/go" "go" "/usr/local/go/bin/go" 0; \
    update-alternatives --set go /usr/local/go/bin/go

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
