#!/usr/bin/env bash
# 自动调用同名 Python 脚本的启动脚本

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../../../.."
export PYTHONPATH="${PROJECT_ROOT}:${PYTHONPATH}"
PYTHON_SCRIPT="${SCRIPT_DIR}/../collect_bmc_log.py"

if [ ! -f "$PYTHON_SCRIPT" ]; then
    echo "错误：找不到对应的 Python 脚本 $PYTHON_SCRIPT" >&2
    exit 1
fi

python3 "$PYTHON_SCRIPT" "$@"