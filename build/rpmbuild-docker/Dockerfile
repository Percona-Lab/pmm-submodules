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
                   epel-release \
                   && \
    yum clean all && rm -rf /var/cache/yum

RUN useradd builder -u 1000 -m -G users,wheel && \
    echo "builder ALL=(ALL:ALL) NOPASSWD:ALL" >> /etc/sudoers && \
    echo "# macros"                      >  /home/builder/.rpmmacros && \
    echo "%_topdir    /home/builder/rpm" >> /home/builder/.rpmmacros && \
    mkdir /home/builder/rpm && \
    chown -R builder /home/builder
USER builder

ENV FLAVOR=rpmbuild OS=centos DIST=el7
WORKDIR /home/builder/rpm
