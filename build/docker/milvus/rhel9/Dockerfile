# Copyright (C) 2024 Red Hat, Inc. All rights reserved.

#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under the License.

FROM registry.access.redhat.com/ubi9/ubi:9.4-947.1717074712

ARG TARGETARCH

RUN dnf install -y wget libgomp libaio libatomic

USER 0

# Assumes you have a valid subman subscription
RUN subscription-manager register --auto-attach
# This gets used for the codeready-builder-for-rhel-9-<arch> stream for openblas-devel 
COPY build/docker/milvus/rhel9/install-openblas.sh /home/install-openblas.sh
RUN chmod +x /home/install-openblas.sh
RUN TARGETARCH=$TARGETARCH /home/install-openblas.sh

# Remove cache for image sizing
RUN rm -rf /var/cache/dnf/*

COPY ./bin/ /milvus/bin/

COPY ./configs/ /milvus/configs/ 

COPY ./lib/ /milvus/lib/

ENV PATH=/milvus/bin:$PATH
ENV LD_LIBRARY_PATH=/milvus/lib:$LD_LIBRARY_PATH:/usr/lib
ENV LD_PRELOAD=/milvus/lib/libjemalloc.so
ENV MALLOC_CONF=background_thread:true

# Add Tini
ADD https://github.com/krallin/tini/releases/download/v0.19.0/tini-$TARGETARCH /tini
RUN chmod +x /tini

USER 1001
ENTRYPOINT ["/tini", "--"]
WORKDIR /milvus
