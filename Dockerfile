# setup build image
FROM golang:1.23.0@sha256:8e529b64d382182bb84f201dea3c72118f6ae9bc01d27190ffc5a54acf0091d2 AS operator-build

RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && apt-get install -y libbtrfs-dev libdevmapper-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

ARG GO_LINKER_ARGS
ARG GO_BUILD_TAGS

COPY pkg ./pkg
COPY cmd ./cmd
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=1 CGO_CFLAGS="-O2 -Wno-return-local-addr" \
    go build -tags "${GO_BUILD_TAGS}" -trimpath -ldflags="${GO_LINKER_ARGS}" \
    -o ./build/_output/bin/dynatrace-operator ./cmd/

FROM registry.access.redhat.com/ubi9-micro:9.4-13@sha256:9dbba858e5c8821fbe1a36c376ba23b83ba00f100126f2073baa32df2c8e183a AS base
FROM registry.access.redhat.com/ubi9:9.4-1181@sha256:1ee4d8c50d14d9c9e9229d9a039d793fcbc9aa803806d194c957a397cf1d2b17 AS dependency
RUN mkdir -p /tmp/rootfs-dependency
COPY --from=base / /tmp/rootfs-dependency
RUN dnf install --installroot /tmp/rootfs-dependency \
      util-linux-core tar \
      --releasever 9 \
      --setopt install_weak_deps=false \
      --nodocs -y \
 && dnf --installroot /tmp/rootfs-dependency clean all \
 && rm -rf \
      /tmp/rootfs-dependency/var/cache/* \
      /tmp/rootfs-dependency/var/log/dnf* \
      /tmp/rootfs-dependency/var/log/yum.*

FROM base

COPY --from=dependency /tmp/rootfs-dependency /

# operator binary
COPY --from=operator-build /app/build/_output/bin /usr/local/bin

# csi binaries
COPY --from=registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.11.1@sha256:e01facb9fb9cffaf52d0053bdb979fbd8c505c8e411939a6e026dd061a6b4fbe /csi-node-driver-registrar /usr/local/bin
COPY --from=registry.k8s.io/sig-storage/livenessprobe:v2.13.1@sha256:d2a9027a4876e039185e9bef7c61a0142c8ea14e7440860285c34ac73fee4ffb /livenessprobe /usr/local/bin

COPY ./third_party_licenses /usr/share/dynatrace-operator/third_party_licenses
COPY LICENSE /licenses/

# custom scripts
COPY hack/build/bin /usr/local/bin

LABEL name="Dynatrace Operator" \
      vendor="Dynatrace LLC" \
      maintainer="Dynatrace LLC" \
      version="1.x" \
      release="1" \
      url="https://www.dynatrace.com" \
      summary="The Dynatrace Operator is an open source Kubernetes Operator for easily deploying and managing Dynatrace components for Kubernetes / OpenShift observability. By leveraging the Dynatrace Operator you can innovate faster with the full potential of Kubernetes / OpenShift and Dynatrace’s best-in-class observability and intelligent automation." \
      description="Automate Kubernetes observability with Dynatrace" \
      io.k8s.description="Automate Kubernetes observability with Dynatrace" \
      io.k8s.display-name="Dynatrace Operator" \
      io.openshift.tags="observability,monitoring,dynatrace,operator,logging,metrics,tracing,prometheus,alerts" \
      vcs-url="https://github.com/Dynatrace/dynatrace-operator.git" \
      vcs-type="git" \
      changelog-url="https://github.com/Dynatrace/dynatrace-operator/releases"

ENV OPERATOR=dynatrace-operator \
    USER_UID=1001 \
    USER_NAME=dynatrace-operator

RUN /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}:${USER_UID}
