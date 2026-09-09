# 合并镜像: 一个 Dockerfile 构建 agent-core (SERVICE=agent-core, HTTP :9700 + gRPC :9710)
# 或 node-collector (SERVICE=node-collector, gRPC :9720)。k8s 部署清单里显式设 SERVICE 复用同一镜像。
FROM python:3.13-slim

ARG SERVICE=agent-core
ENV SERVICE=${SERVICE}

WORKDIR /app

# dmidecode 供 node-collector 采集 SMBIOS; passwd 提供 useradd/usermod
RUN apt-get update && apt-get install -y --no-install-recommends dmidecode passwd \
    && rm -rf /var/lib/apt/lists/*

# 与 clusterd 一致: 运行用户 hwMindX (uid 9000), 禁止 root/hwMindX 交互登录
RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX && \
    usermod root -s /usr/sbin/nologin

# fd + 组件 whl: 全部复制到 hwMindX 家目录, 按架构挑选 fd wheel 后一次安装
COPY ascend_faultdiag-*.whl agent_core-*.whl node_collector-*.whl /home/hwMindX/
RUN _arch="$(uname -m)"; \
    [ "$_arch" = "aarch64" ] || _arch="x86_64"; \
    _fd_whl="$(ls /home/hwMindX/ascend_faultdiag-*-linux_${_arch}.whl 2>/dev/null | head -n1)"; \
    if [ -z "$_fd_whl" ]; then \
        echo "ERROR: no ascend_faultdiag-*-linux_${_arch}.whl next to Dockerfile (see README.md)"; \
        exit 1; \
    fi; \
    pip install --no-cache-dir "$_fd_whl" /home/hwMindX/agent_core-*.whl /home/hwMindX/node_collector-*.whl

# 采集契约内置为镜像默认值 (运行期可被 cluster-system/collect-manifest CM 覆盖)
COPY collect_manifest.yaml /home/hwMindX/collect_manifest.yaml
RUN chown hwMindX:hwMindX /home/hwMindX/collect_manifest.yaml && \
    chmod 440 /home/hwMindX/collect_manifest.yaml

EXPOSE 9700 9710 9720

# hwMindX 为镜像默认运行用户 (非特权)
USER hwMindX

# 启动: 按 SERVICE 参数启动对应项目
CMD case "${SERVICE:-agent-core}" in \
        node-collector) exec python -m node_collector.collector ;; \
        agent-core)     exec uvicorn agent_core.main:app --host 0.0.0.0 --port "${AGENT_PORT:-9700}" ;; \
        *) echo "unknown SERVICE=${SERVICE} (expect agent-core|node-collector)" >&2; exit 1 ;; \
    esac
