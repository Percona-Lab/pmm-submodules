FROM centos:7

# enable nodesource repo for nodejs
RUN curl -sL https://rpm.nodesource.com/setup_12.x | bash -
RUN yum update -y
RUN yum install -y --setopt=skip_missing_names_on_install=False \
                   gcc gcc-c++ \
                   nodejs \
                   libtool libtool-ltdl \
                   make cmake \
                   git \
                   pkgconfig \
                   sudo \
                   automake autoconf \
                   rpmdevtools createrepo_c epel-release \
                   yum-utils rpm-build \
                   wget \
                   glibc-static
RUN yum -y remove nodesource-release-el7-1.noarch
RUN yum clean all && rm -rf /var/cache/yum

ENV GO_TARBALL https://dl.google.com/go/go1.12.12.linux-amd64.tar.gz
RUN wget --progress=dot:giga ${GO_TARBALL} -O /tmp/golang.tar.gz
RUN tar -C /usr/local -xzf /tmp/golang.tar.gz; \
    update-alternatives --install "/usr/bin/go" "go" "/usr/local/go/bin/go" 0; \
    update-alternatives --set go /usr/local/go/bin/go
RUN rm /tmp/golang.tar.gz

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
