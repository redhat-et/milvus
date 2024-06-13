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

if [[ "$RUN_MODE" == "aarch64" ]] || [[ "$RUN_MODE" == "x86_64" ]]; then
    dnf install -y --enablerepo=codeready-builder-for-rhel-9-$RUN_MODE-rpms openblas-devel
else 
    echo "uncaught runmode based on invalid \$TARGETARCH."
    exit 1
fi
