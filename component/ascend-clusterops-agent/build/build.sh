#!/usr/bin/env bash
# ============================================================================
# 构建 agent-core / node-collector 为 whl 包, 并组装部署/构建文件到 output/ 目录
#
# 运行后产物:
#   build/dist/         两个组件 whl (agent_core / node_collector)
#   <组件根目录>/output/  交付物目录, 由流水线 build_package.sh 统一打包 zip
#
# output/ 内容:
#   agent_core-*.whl
#   node_collector-*.whl
#   agent-core.yaml
#   node-collector.yaml
#   Dockerfile
#   collect_manifest.yaml
#   kubectl-plugin/                     (install.sh + kubectl-ascend_diag + kubectl-clusterops)
#
# 注意事项:
#   - 版本号默认 v26.2.0; 若组件根目录存在 service_config.ini, 则读取其中版本号
#     (格式同其他发行包: 首行 key=value, 取 '=' 后的内容, 最终形如 v26.2.0)。
#   - ascend_faultdiag-*.whl 不打包进 zip。构建镜像前需用户自行下载, 放到
#     Dockerfile 同目录 (仓库内为 build/, 交付包解压后为解压目录)。
# ============================================================================
set -euo pipefail

CUR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPONENT_DIR="$(realpath "${CUR_DIR}/..")"
PY="${PY:-python3}"

VER_FILE="${COMPONENT_DIR}"/service_config.ini
build_version="v26.2.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  # 取 '=' 后的内容作为版本号, 最终形如 v26.2.0
  build_version="v"${line#*=}
fi
# whl 中间产物目录 (复制进 output/)
WHL_DIR="${CUR_DIR}/dist"
# 交付物输出目录 (流水线 build_package.sh 约定: component/<name>/output/)
OUTPUT_DIR="${COMPONENT_DIR}/output"

function clean() {
    rm -rf "${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
}

# 检查 ascend-fd 安装包是否已提前就绪 (build.sh 不下载, 见 README.md 下载地址)
function check_ascend_fd() {
    if compgen -G "${CUR_DIR}/ascend_faultdiag-*.whl" >/dev/null; then
        echo "==> ascend-fd whl 已就绪:"
        ls -1 "${CUR_DIR}"/ascend_faultdiag-*.whl
        return
    fi
    echo "WARN: ${CUR_DIR}/ 下没有 ascend_faultdiag-*.whl, 构建镜像时 ascend-fd 将无法安装!"
    echo "      请提前下载 Ascend-mindxdl-faultdiag_<ver>_linux-<arch>.zip, 仅提取"
    echo "      其中 ascend_faultdiag-*.whl 放到 Dockerfile 同目录 ${CUR_DIR}/ (下载地址见 README.md)"
}

function build_whl() {
    cd "${COMPONENT_DIR}"
    # 每次从干净状态构建, 避免多版本 whl 累积
    rm -rf "${WHL_DIR}"
    mkdir -p "${WHL_DIR}"

    # whl 版本号与 zip 交付版本号保持一致 (BUILD_VERSION 覆盖时同步注入; 构建后还原源码)
    _pyprojects=("${COMPONENT_DIR}/agent-core/pyproject.toml" "${COMPONENT_DIR}/node-collector/pyproject.toml")
    _whl_version="${build_version#v}"
    for f in "${_pyprojects[@]}"; do
        cp "$f" "$f.orig"
        sed -i "s/^version *= *\"[^\"]*\"/version = \"${_whl_version}\"/" "$f"
    done
    trap 'for f in "${_pyprojects[@]}"; do [ -f "$f.orig" ] && mv -f "$f.orig" "$f"; done' EXIT

    echo "==> [1/2] build agent-core =="
    "${PY}" -m pip wheel --no-deps --no-build-isolation --wheel-dir "${WHL_DIR}" ./agent-core

    echo "==> [2/2] build node-collector =="
    "${PY}" -m pip wheel --no-deps --no-build-isolation --wheel-dir "${WHL_DIR}" ./node-collector

    echo
    echo "==> whl 产物 (${WHL_DIR}) =="
    ls -1 "${WHL_DIR}"/*.whl
}

function package() {
    # 组件 whl (具体以编译产物为准)
    cp "${WHL_DIR}"/agent_core-*.whl "${OUTPUT_DIR}"/
    cp "${WHL_DIR}"/node_collector-*.whl "${OUTPUT_DIR}"/

    # 部署/构建文件
    cp "${CUR_DIR}"/agent-core.yaml "${OUTPUT_DIR}"/
    cp "${CUR_DIR}"/node-collector.yaml "${OUTPUT_DIR}"/
    cp "${CUR_DIR}"/Dockerfile "${OUTPUT_DIR}"/
    cp "${CUR_DIR}"/collect_manifest.yaml "${OUTPUT_DIR}"/

    # kubectl 插件 (kubectl-plugin/ 子目录; install.sh 源在 build/, 打包时并入)
    mkdir -p "${OUTPUT_DIR}"/kubectl-plugin
    cp "${COMPONENT_DIR}"/kubectl-plugin/kubectl-ascend_diag "${OUTPUT_DIR}"/kubectl-plugin/
    cp "${COMPONENT_DIR}"/kubectl-plugin/kubectl-clusterops "${OUTPUT_DIR}"/kubectl-plugin/
    cp "${CUR_DIR}"/install.sh "${OUTPUT_DIR}"/kubectl-plugin/install.sh
    chmod +x "${OUTPUT_DIR}"/kubectl-plugin/install.sh \
        "${OUTPUT_DIR}"/kubectl-plugin/kubectl-ascend_diag \
        "${OUTPUT_DIR}"/kubectl-plugin/kubectl-clusterops

    echo
    echo "==> 交付物目录: ${OUTPUT_DIR} =="
    ls -1 "${OUTPUT_DIR}"
    ls -1 "${OUTPUT_DIR}"/kubectl-plugin
}

function main() {
    clean
    check_ascend_fd
    build_whl
    package
}

main
