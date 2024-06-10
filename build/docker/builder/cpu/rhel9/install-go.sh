#/bin/bash
if [[ -z "$GOROOT" ]]; then
    GOROOT=/usr/local/go 
fi

mkdir -p $GOROOT

RUN_MODE=""
if [[ -z "$TARGETARCH" ]]; then
    TARGETARCH=$(uname -m)
    echo "\$TARGETARCH not set - defaulting to host system"
fi

if [[ "$TARGETARCH" == "arm64" ]] || [[ "$TARGETARCH" == "aarch64" ]]; then
    RUN_MODE="arm64"
elif [[ "$TARGETARCH" == "amd64" ]] || [[ "$TARGETARCH" == "x86_64" ]]; then
    RUN_MODE="amd64"
else
    echo "invalid \$TARGETARCH, build failure.
    Supported options: ['arm64', 'aarch64', 'amd64', 'x86_64']"
    exit 1
fi

export GO111MODULE="on"

if [[ -z "$GOPATH" ]]; then
    export GOPATH="/go"
fi

if [[ "$RUN_MODE" == "arm64" ]] || [[ "$RUN_MODE" == "amd64" ]] ; then
    wget -qO- "https://go.dev/dl/go1.21.10.linux-$RUN_MODE.tar.gz" | tar --strip-components=1 -xz -C $GOROOT
else 
    echo "uncaught \$RUN_MODE based on invalid \$TARGETARCH."
    exit 1
fi

export PATH="$GOPATH/bin:$GOROOT/bin:$PATH"

mkdir -p "$GOPATH/src" "$GOPATH/bin"
go clean --modcache
chmod -R 777 "$GOPATH" && chmod -R a+w $(go env GOTOOLDIR)