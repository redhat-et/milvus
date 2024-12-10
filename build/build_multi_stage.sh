#!/bin/bash

# Licensed to the LF AI & Data foundation under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

DOCKERFILE="${DOCKERFILE:-./build/multi-stage/cpu/ubi-rhel/Dockerfile}"
IMAGE_ARCH="${IMAGE_ARCH:-amd64}"
MILVUS_IMAGE_REPO="${MILVUS_IMAGE_REPO:-quay.io/redhat-et/milvus}"
MILVUS_IMAGE_TAG="${MILVUS_IMAGE_TAG:-standalone-rhel9.5}"
CONTAINER_CMD="${CONTAINER_CMD:-podman}"

# RH distros package repos use x86_64 rather than amd64, and aarch64 rather than arm64
RH_TARGETARCH="${RH_TARGETARCH:-x86_64}"

if [ -z "$IMAGE_ARCH" ]; then
    MACHINE=$(uname -m)
    if [ "$MACHINE" = "x86_64" ]; then
        IMAGE_ARCH="amd64"
        RH_TARGETARCH="x86_64"
    else
        IMAGE_ARCH="arm64"
        RH_TARGETARCH="aarch64"
    fi
fi

echo ${IMAGE_ARCH}

BUILD_ARGS="${BUILD_ARGS:---build-arg TARGETARCH=${IMAGE_ARCH} --build-arg RH_TARGETARCH=${RH_TARGETARCH}}"

${CONTAINER_CMD} build --network host ${BUILD_ARGS} --platform linux/${IMAGE_ARCH} -f "${DOCKERFILE}" -t "${MILVUS_IMAGE_REPO}:${MILVUS_IMAGE_TAG}" .
