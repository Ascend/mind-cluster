#!/usr/bin/env bash
# 执行 ascend-clusterops-agent 全部单元测试 (pytest; conftest 自动挂载共享包
# clusterops_common / diagproto / agent_core / node_collector)
# 用法: bash tests/run_dt.sh [pytest 参数...]   (不带参数时默认 -q)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPONENT_DIR="$(realpath "${SCRIPT_DIR}/..")"
PY="${PY:-python3}"

cd "${COMPONENT_DIR}"
if [ "$#" -eq 0 ]; then
    exec "${PY}" -m pytest tests/ -q
fi
exec "${PY}" -m pytest tests/ "$@"
