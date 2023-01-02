# This Dockerfile is used to build the image available on DockerHub
FROM golang:1.18 as build

# Add everything
ADD . /usr/src/cni-route-override

RUN /bin/bash /usr/src/cni-route-override/build_linux.sh

FROM alpine
LABEL org.opencontainers.image.source https://github.com/redhat-nfvpe/cni-route-override
COPY --from=build /usr/src/cni-route-override/bin/route-override /
WORKDIR /
