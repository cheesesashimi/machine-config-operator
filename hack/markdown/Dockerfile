# Installing googleworkspace-cli requires tar which is not included in the base image.
FROM quay.io/hummingbird/nodejs:25-builder AS install-tar
USER 0
RUN dnf install -y tar gzip && dnf clean all

FROM quay.io/hummingbird/nodejs:25-builder AS install-git
USER 0
RUN mkdir -p /staged && \
    dnf install -y --use-host-config --installroot=/staged \
      --setopt=install_weak_deps=False --nodocs git && \
    rm -rf /staged/var/lib/dnf /staged/var/log/* /staged/var/lib/rpm

FROM quay.io/hummingbird/nodejs:25 AS install-node-packages
ENV NPM_CONFIG_PREFIX=/tmp/npm-global
ENV PATH=$NPM_CONFIG_PREFIX/bin:$PATH
WORKDIR /workdir
COPY --from=install-tar /usr /usr
COPY install-markdownlint.sh /tmp
RUN /tmp/install-markdownlint.sh

FROM quay.io/hummingbird/nodejs:25 AS final
ENV NPM_CONFIG_PREFIX=/tmp/npm-global
ENV PATH=$NPM_CONFIG_PREFIX/bin:$PATH
COPY --from=install-node-packages /tmp/npm-global /tmp/npm-global
COPY --from=install-git /staged/usr /usr
WORKDIR /workdir
ENTRYPOINT /workdir/hack/markdown/markdownlint.sh
