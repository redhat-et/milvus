## HOW TO BUILD

### Build Milvus image

The following environment variables can be set for the build, shown here with defaults:

```
MILVUS_IMAGE_REPO="${MILVUS_IMAGE_REPO:-quay.io/redhat-et/milvus}"
MILVUS_IMAGE_TAG="${MILVUS_IMAGE_TAG:-standalone-rhel9.5}"
IMAGE_ARCH="${IMAGE_ARCH:-amd64}"
RH_TARGETARCH="${RH_TARGETARCH:-x86_64}"
BUILD_ARGS="${BUILD_ARGS:---build-arg TARGETARCH=${IMAGE_ARCH} --build-arg RH_TARGETARCH=${RH_TARGETARCH}}"
```

Run the following from the root of the repository.

```bash
./build/build_multi_stage.sh
```

The build takes a long time depending on your system, ~1hr

## HOW TO RUN

### Run standalone milvus service

Create the milvus data volume

```bash
mkdir -p /tmp/volumes/milvus
```

Run milvus standalone

```bash
podman run --rm -d \
    --name milvus-standalone \
    --security-opt seccomp:unconfined \
    -v /tmp/volumes/milvus:/var/lib/milvus:Z  \
    -p 19530:19530 \
    -p 9091:9091 \
    -p 2379:2379 \
    -e ETCD_DATA_DIR=/var/lib/milvus/etcd \
    --health-cmd="curl -f http://localhost:9091/healthz" \
    --health-interval=30s \
    --health-start-period=90s \
    --health-timeout=20s \
    --health-retries=3 \
    quay.io/redhat-et/milvus:standalone-rhel9.5 \
    milvus run standalone
```

### Run RAG application with milvus

Follow [this example](../../../deployments/podman-kube-play/README.md) to deploy a local RAG application with standalone Milvus.
