FROM registry.access.redhat.com/ubi9/ubi:9.4-947.1717074712 as vcpkg-installer
USER 0
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

########################################################################################
FROM registry.access.redhat.com/ubi9/ubi:9.4-947.1717074712

ARG TARGETARCH
ARG SUBMAN_USER
ARG SUBMAN_PASS
USER 0

# basic deps
RUN dnf install -y make cmake automake gcc gcc-c++ \
        zip git libaio libuuid-devel wget libstdc++-static \
        python3.11 python3.11-pip libatomic libtool ninja-build \
        perl-IPC-Cmd perl-Digest-SHA perl-FindBin 

RUN alias python3='python3.11'

# Assumes you have a valid subman subscription
    # This gets used for the codeready-builder-for-rhel-9-<arch> stream for openblas-devel

COPY build/docker/builder/cpu/rhel9/install-rpms.sh /root/install-rpms.sh
RUN chmod +x /root/install-rpms.sh
RUN /root/install-rpms.sh

# Install conan and Go
RUN python3.11 -m pip install conan==1.61.0

COPY build/docker/builder/cpu/rhel9/install-go.sh /root/install-go.sh
RUN chmod +x /root/install-go.sh
RUN source /root/install-go.sh

RUN mkdir -p /.cache/go-build && chmod -R 777 /.cache/go-build

RUN curl https://sh.rustup.rs -sSf | \
    sh -s -- --default-toolchain=1.73 -y

ENV PATH=/root/.cargo/bin:/usr/local/bin:/usr/local/go/bin:$PATH

ENV VCPKG_FORCE_SYSTEM_BINARIES 1

RUN mkdir /opt/vcpkg &&  \
    wget -qO- vcpkg.tar.gz https://github.com/microsoft/vcpkg/archive/master.tar.gz | tar --strip-components=1 -xz -C /opt/vcpkg && \
    rm -rf vcpkg.tar.gz

COPY --from=vcpkg-installer /root/.cache/vcpkg /root/.cache/vcpkg
COPY --chown=0:0 build/docker/builder/entrypoint.sh /

ENTRYPOINT [ "/entrypoint.sh" ]
CMD ["tail", "-f", "/dev/null"]
