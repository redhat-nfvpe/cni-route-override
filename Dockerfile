# This Dockerfile is used to build the image available on DockerHub
FROM centos:centos7 as build

# Add everything
ADD . /usr/src/cni-route-override

ENV INSTALL_PKGS "git golang"
RUN rpm --import https://mirror.go-repo.io/centos/RPM-GPG-KEY-GO-REPO && \
    curl -s https://mirror.go-repo.io/centos/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo && \
    yum install -y $INSTALL_PKGS && \
    rpm -V $INSTALL_PKGS && \
    cd /usr/src/cni-route-override && \
    ./build_linux.sh

FROM alpine
LABEL org.opencontainers.image.source https://github.com/redhat-nfvpe/cni-route-override
COPY --from=build /usr/src/cni-route-override/bin/route-override /
WORKDIR /
