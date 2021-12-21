FROM centos:7

# enable nodesource repo for nodejs
RUN curl -sL https://rpm.nodesource.com/setup_14.x | bash -
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

# keep that format for easier search
ENV GO_VERSION go1.17.5
ENV GO_CHECKSUM bd78114b0d441b029c8fe0341f4910370925a4d270a6a590668840675b0c653e
ENV GO_RELEASER_CHECKSUM db351ae63bdb9f3834a9a76caa5845b4de4c403855a6524afb1d2a97f82e627f

RUN wget --progress=dot:giga https://dl.google.com/go/${GO_VERSION}.linux-amd64.tar.gz -O /tmp/golang.tar.gz
RUN wget --progress=dot:giga https://github.com/goreleaser/goreleaser/releases/download/v0.142.0/goreleaser_amd64.rpm -O /tmp/goreleaser.rpm
RUN sha256sum /tmp/golang.tar.gz
RUN echo "${GO_CHECKSUM} /tmp/golang.tar.gz" | sha256sum -c -
RUN sha256sum /tmp/goreleaser.rpm
RUN echo "${GO_RELEASER_CHECKSUM} /tmp/goreleaser.rpm" | sha256sum -c -
RUN tar -C /usr/local -xzf /tmp/golang.tar.gz
RUN yum install -y /tmp/goreleaser.rpm
RUN update-alternatives --install "/usr/bin/go" "go" "/usr/local/go/bin/go" 0
RUN update-alternatives --set go /usr/local/go/bin/go
RUN update-alternatives --install "/usr/bin/gofmt" "gofmt" "/usr/local/go/bin/gofmt" 0
RUN update-alternatives --set gofmt /usr/local/go/bin/gofmt
RUN update-alternatives --install "/usr/bin/goreleaser" "goreleaser" "/usr/local/goreleaser" 0
RUN update-alternatives --set goreleaser /usr/local/goreleaser
RUN rm /tmp/golang.tar.gz /tmp/goreleaser.rpm

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
