FROM centos:7

# enable nodesource repo for nodejs
RUN curl -sL https://rpm.nodesource.com/setup_12.x | bash -
RUN curl -sL https://dl.yarnpkg.com/rpm/yarn.repo | tee /etc/yum.repos.d/yarn.repo
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
                   glibc-static \
                   yarn
RUN yum -y remove nodesource-release-el7-1.noarch
RUN yum clean all && rm -rf /var/cache/yum

ENV GO_VERSION 1.13.10
ENV GO_CHECKSUM 8a4cbc9f2b95d114c38f6cbe94a45372d48c604b707db2057c787398dfbf8e7f

RUN wget --progress=dot:giga https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -O /tmp/golang.tar.gz
RUN sha256sum /tmp/golang.tar.gz
RUN echo "${GO_CHECKSUM} /tmp/golang.tar.gz" | sha256sum -c -
RUN tar -C /usr/local -xzf /tmp/golang.tar.gz
RUN update-alternatives --install "/usr/bin/go" "go" "/usr/local/go/bin/go" 0
RUN update-alternatives --set go /usr/local/go/bin/go
RUN update-alternatives --install "/usr/bin/gofmt" "gofmt" "/usr/local/go/bin/gofmt" 0
RUN update-alternatives --set gofmt /usr/local/go/bin/gofmt
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
