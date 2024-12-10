## Example RAG application with standalone Milvus

This assumes you have a local model-server running and the MODEL_ENDPOINT is `http://127.0.0.1:8001`. Adjust [rag-milvus-ex.yaml](./rag-milvus-ex.yaml) as necessary,
according to your environment.

For an example of a simple llamacpp model-server,
see [github.com/containers/ai-lab-recipes](https://github.com/containers/ai-lab-recipes/tree/main/model_servers/llamacpp_python#deploy-model-service).

### Deploy RAG application

Create the milvus data volume on the local file system.

```bash
mkdir -p /tmp/volumes/milvus
```

Run podman rag-milvus pod

```bash
podman kube play ./rag-milvus-ex.yaml
```

### Interact with RAG application

Visit local browser at `http://localhost:8501`

### View pod logs

```bash
podman pod list
podman pod logs rag-app-milvus
```

### Stop RAG application

```bash
podman kube down ./rag-milvus-ex.yaml
```
