#/bin/bash
RUN_MODE=""
if [[ -z "$TARGETARCH" ]]; then
    TARGETARCH=$(uname -m)
    echo "\$TARGETARCH not set - defaulting to host system"
fi

if [[ "$TARGETARCH" == "arm64" ]] || [[ "$TARGETARCH" == "aarch64" ]]; then
    RUN_MODE="aarch64"
elif [[ "$TARGETARCH" == "amd64" ]] || [[ "$TARGETARCH" == "x86_64" ]]; then
    RUN_MODE="x86_64"
else
    echo "invalid \$TARGETARCH, build failure.
    Supported options: ['arm64', 'aarch64', 'amd64', 'x86_64']"
    exit 1
fi

subscription-manager attach

if [[ "$RUN_MODE" == "aarch64" ]]; then
    dnf install -y \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/aarch64/os/Packages/perl-Unicode-EastAsianWidth-12.0-7.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/aarch64/os/Packages/perl-Text-Unidecode-1.30-16.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/aarch64/os/Packages/perl-libintl-perl-1.32-4.el9.aarch64.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/CRB/aarch64/os/Packages/texinfo-6.7-15.el9.aarch64.rpm
    dnf install -y --enablerepo=codeready-builder-for-rhel-9-aarch64-rpms openblas-devel
elif [[ "$RUN_MODE" == "x86_64" ]]; then
    dnf install -y \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/x86_64/os/Packages/perl-Unicode-EastAsianWidth-12.0-7.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/x86_64/os/Packages/perl-Text-Unidecode-1.30-16.el9.noarch.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/AppStream/x86_64/os/Packages/perl-libintl-perl-1.32-4.el9.x86_64.rpm \
        https://www.rpmfind.net/linux/centos-stream/9-stream/CRB/x86_64/os/Packages/texinfo-6.7-15.el9.x86_64.rpm
    dnf install -y --enablerepo=codeready-builder-for-rhel-9-x86_64-rpms openblas-devel
else 
    echo "uncaught runmode based on invalid \$TARGETARCH."
    exit 1
fi
