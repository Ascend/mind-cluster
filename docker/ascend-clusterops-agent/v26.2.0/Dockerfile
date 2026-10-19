# Ascend ClusterOps Agent merged image: one Dockerfile builds agent-core
# (SERVICE=agent-core, HTTP 9700 + gRPC 9710) or node-collector
# (SERVICE=node-collector, gRPC 9720). Set SERVICE explicitly in k8s
# manifests to reuse the same image.
FROM python:3.13-slim AS build

ARG VERSION=26.2.0
ARG TARGETPLATFORM

RUN apt update && apt install -y wget unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/tmp/* \
    && rm -rf /tmp/*

# Download the arch-independent agent package and the arch-specific faultdiag
# package, then keep only the faultdiag whl (pip install needs internet access)
RUN ARCH=$(case "${TARGETPLATFORM}" in \
        "linux/amd64") echo "x86_64" ;; \
        "linux/arm64") echo "aarch64" ;; \
        *) echo "Unsupported TARGETPLATFORM: ${TARGETPLATFORM}" && exit 1 ;; \
    esac) &&\
    wget https://gitcode.com/ascend/mind-cluster/releases/download/v${VERSION}/Ascend-mindxdl-ascend-clusterops-agent_${VERSION}_linux.zip &&\
    unzip Ascend-mindxdl-ascend-clusterops-agent_${VERSION}_linux.zip -d /build &&\
    wget https://gitcode.com/ascend/mind-cluster/releases/download/v${VERSION}/Ascend-mindxdl-faultdiag_${VERSION}_linux-${ARCH}.zip &&\
    unzip Ascend-mindxdl-faultdiag_${VERSION}_linux-${ARCH}.zip ascend_faultdiag-*.whl -d /build

FROM python:3.13-slim

ARG SERVICE=agent-core
ENV SERVICE=${SERVICE}

WORKDIR /app

# dmidecode: node-collector SMBIOS collection; passwd: useradd/usermod
RUN apt-get update && apt-get install -y --no-install-recommends dmidecode passwd \
    && rm -rf /var/lib/apt/lists/*

# Same as clusterd: run as hwMindX (uid 9000), interactive login disabled
RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

# Install the faultdiag whl and the component whls; runtime deps are pulled
# from PyPI by pip (requires internet access during build)
COPY --from=build /build/ascend_faultdiag-*.whl \
    /build/agent_core-*.whl \
    /build/node_collector-*.whl \
    /home/hwMindX/
RUN pip install --no-cache-dir /home/hwMindX/ascend_faultdiag-*.whl \
    /home/hwMindX/agent_core-*.whl \
    /home/hwMindX/node_collector-*.whl

# Built-in collect manifest as image default (overridable at runtime by the
# cluster-system/collect-manifest ConfigMap)
COPY --from=build /build/collect_manifest.yaml /home/hwMindX/collect_manifest.yaml
RUN chown hwMindX:hwMindX /home/hwMindX/collect_manifest.yaml &&\
    chmod 440 /home/hwMindX/collect_manifest.yaml

EXPOSE 9700 9710 9720

USER hwMindX

# Start the service selected by the SERVICE arg
CMD case "${SERVICE:-agent-core}" in \
        node-collector) exec python -m node_collector.collector ;; \
        agent-core)     exec uvicorn agent_core.main:app --host 0.0.0.0 --port "${AGENT_PORT:-9700}" ;; \
        *) echo "unknown SERVICE=${SERVICE} (expect agent-core|node-collector)" >&2; exit 1 ;; \
    esac
