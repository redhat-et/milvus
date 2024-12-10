FROM registry.access.redhat.com/ubi9:9.5-1732804088 as vcpkg-installer

RUN dnf install -y wget zip git gcc gcc-c++ cmake \
        dnf-plugins-core ninja-build \
        perl-IPC-Cmd perl-Digest-SHA \
        perl-FindBin perl-File-Compare perl-File-Copy 

ENV VCPKG_FORCE_SYSTEM_BINARIES 1

RUN mkdir /opt/vcpkg &&  \
    wget -qO- vcpkg.tar.gz https://github.com/microsoft/vcpkg/archive/refs/tags/2023.11.20.tar.gz | tar --strip-components=1 -xz -C /opt/vcpkg && \
    rm -rf vcpkg.tar.gz

# empty the vscpkg toolchains linux.cmake file to avoid the error
RUN echo "" > /opt/vcpkg/scripts/toolchains/linux.cmake

# install azure-identity-cpp azure-storage-blobs-cpp gtest via vcpkg
RUN /opt/vcpkg/bootstrap-vcpkg.sh -disableMetrics && \
    ln -s /opt/vcpkg/vcpkg /usr/local/bin/vcpkg && \
    vcpkg version && \
    vcpkg install azure-identity-cpp azure-storage-blobs-cpp gtest


#FROM nvcr.io/nvidia/cuda:12.4.1-devel-ubi9
FROM registry.access.redhat.com/ubi9:9.5-1732804088 as builder

ARG RH_TARGETARCH
ARG TARGETARCH

RUN dnf update -y && dnf install -y \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/$RH_TARGETARCH/os/Packages/perl-Unicode-EastAsianWidth-12.0-7.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/$RH_TARGETARCH/os/Packages/perl-Text-Unidecode-1.30-16.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/$RH_TARGETARCH/os/Packages/perl-libintl-perl-1.32-4.el9.$RH_TARGETARCH.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/CRB/$RH_TARGETARCH/os/Packages/texinfo-6.7-15.el9.$RH_TARGETARCH.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/$RH_TARGETARCH/os/Packages/libaio-devel-0.3.111-13.el9.$RH_TARGETARCH.rpm \
        https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm

RUN dnf install -y \
    git \
    gcc \
    gcc-c++ \
    glibc-devel \
    gdb \
    make \
    #cmake \
    automake \
    libffi-devel \
    libtool \
    pkgconfig \
    wget \
    unzip \
    tar \
    bzip2 \
    openssl-devel \
    libaio \
    libstdc++-static \
    llvm-toolset \
    ninja-build \
    openblas \
    openblas-devel \
    zip unzip \
    zlib-devel \
    libcurl-devel \
    libuuid-devel \
    perl-IPC-Cmd perl-Digest-SHA perl-FindBin \
    python3-pip && \
    pip3 install --upgrade pip && \
    pip3 install conan==1.64.1 && \
    dnf clean all && \
    rm -rf /var/cache/dnf

# Install cmake version 3.27, upstream uses this version, and the 'dnf install cmake' version fails
RUN wget -qO- "https://cmake.org/files/v3.27/cmake-3.27.5-linux-`uname -m`.tar.gz" | tar --strip-components=1 -xz -C /usr/local

WORKDIR /home/milvus

# TODO: switch to main branch
RUN git clone --depth 1 --branch rh-build https://github.com/sallyom/milvus.git . \
    && rm -rf .git

COPY --chown=1001:1001 build/docker/builder/entrypoint.sh /home/milvus/

COPY --from=vcpkg-installer /root/.cache/vcpkg /home/milvus/.cache/vcpkg

# Install go
RUN mkdir -p /usr/local/go && wget -qO- "https://go.dev/dl/go1.23.3.linux-$TARGETARCH.tar.gz" | tar --strip-components=1 -xz -C /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

# Install rust
RUN curl https://sh.rustup.rs -sSf | \
    sh -s -- --default-toolchain=1.73 -y
ENV PATH=/home/milvus/.cargo/bin:/usr/local/bin:/usr/local/go/bin:$PATH

# Option - Add Include Path instead of symlinks below?
#ENV CMAKE_INCLUDE_PATH=/usr/include/openblas:$CMAKE_INCLUDE_PATH
RUN for file in /usr/include/openblas/*; do \
      ln -s "$file" /usr/include/"$(basename "$file")"; \
    done

ENV HOME=/home/milvus
RUN useradd -m milvus && mkdir -p /home/milvus/conan && chown -R milvus:milvus /home/milvus
RUN chown -R milvus:milvus /home/milvus
USER milvus
RUN /home/milvus/entrypoint.sh

# configs are for standalone with embedded ETCD
RUN cp /home/milvus/build/multi-stage/cpu/ubi-rhel/configs/* /home/milvus/configs/.

ENV CONAN_USER_HOME=/home/milvus/conan
RUN make install

FROM registry.access.redhat.com/ubi9:9.5-1732804088

ARG TARGETARCH
ARG RH_TARGETARCH

RUN dnf update -y && dnf install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm
RUN dnf install -y \
    https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/$RH_TARGETARCH/os/Packages/libaio-devel-0.3.111-13.el9.$RH_TARGETARCH.rpm \
    ca-certificates \
    libaio \
    libgomp \
    openblas \
    openblas-devel && \
    dnf clean all && \
    rm -rf /var/cache/dnf

WORKDIR /milvus

RUN curl -L -o /milvus/tini https://github.com/krallin/tini/releases/download/v0.19.0/tini-$TARGETARCH && \
    chmod +x /milvus/tini

COPY --chmod=774 --from=builder /home/milvus/bin/ /milvus/bin/
COPY --chmod=774 --from=builder /home/milvus/configs/ /milvus/configs/
COPY --chmod=774 --from=builder /home/milvus/lib/ /milvus/lib/

ENV PATH=/milvus/bin:$PATH
ENV LD_LIBRARY_PATH=/milvus/lib:$LD_LIBRARY_PATH:/usr/lib
ENV LD_PRELOAD=/milvus/lib/libjemalloc.so
ENV MALLOC_CONF=background_thread:true
ENV ETCD_USE_EMBED=true
ENV ETCD_CONFIG_PATH=/milvus/configs/embedEtcd.yaml
ENV COMMON_STORAGETYPE=local

ENTRYPOINT ["/milvus/tini", "--"]
