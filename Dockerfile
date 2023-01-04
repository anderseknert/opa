# Copyright 2019 The OPA Authors.  All rights reserved.
# Use of this source code is governed by an Apache2
# license that can be found in the LICENSE file.

ARG BASE

FROM ${BASE}

LABEL org.opencontainers.image.authors="Torin Sandall <torinsandall@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/open-policy-agent/opa"

# Temporarily allow us to identify whether running from within an offical
# Docker image, so that we may print a warning when uid or gid == 0 (root)
# Remove once https://github.com/open-policy-agent/opa/issues/4295 is done
ENV OPA_DOCKER_IMAGE="official"

ARG OPA_ROOTLESS_IMAGE="false"
ENV OPA_ROOTLESS_IMAGE=${OPA_ROOTLESS_IMAGE}

ARG USER=1000:1000
USER ${USER}

# TARGETOS and TARGETARCH are automatic platform args injected by BuildKit
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETOS
ARG TARGETARCH
ARG BIN_DIR=.
ARG BIN_SUFFIX=
COPY ${BIN_DIR}/opa_${TARGETOS}_${TARGETARCH}${BIN_SUFFIX} /opa
ENV PATH=${PATH}:/

ENTRYPOINT ["/opa"]
CMD ["run"]
