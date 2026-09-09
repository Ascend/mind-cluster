#!/usr/bin/env bash
# ============================================================================
# 构建 agent-core / node-collector 为 whl 包, 并与部署/构建文件打包为 zip 交付物
#
# 运行后产物:
#   build/dist/                                         两个组件 whl (agent_core / node_collector)
#   <组件根目录>/Ascend-mindxdl-ascend-clusterops-agent_<ver>_linux.zip   交付包
#
# 交付包 (zip) 内容:
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
output_name="Ascend-mindxdl-ascend-clusterops-agent"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  # 取 '=' 后的内容作为版本号, 最终形如 v26.2.0
  build_version="v"${line#*=}
fi
# zip 包名规范: Ascend-mindxdl-<name>_<ver>_linux.zip (linux 平台族, 不细分架构)
zip_name="${output_name}_${build_version#v}_linux.zip"

# whl 中间产物目录 (打包进 zip)
WHL_DIR="${CUR_DIR}/dist"
# zip 打包暂存目录
PACK_DIR="${CUR_DIR}/package"

function clean() {
    rm -rf "${PACK_DIR}"
    mkdir -p "${PACK_DIR}"
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
    cp "${WHL_DIR}"/agent_core-*.whl "${PACK_DIR}"/
    cp "${WHL_DIR}"/node_collector-*.whl "${PACK_DIR}"/

    # 部署/构建文件
    cp "${CUR_DIR}"/agent-core.yaml "${PACK_DIR}"/
    cp "${CUR_DIR}"/node-collector.yaml "${PACK_DIR}"/
    cp "${CUR_DIR}"/Dockerfile "${PACK_DIR}"/
    cp "${CUR_DIR}"/collect_manifest.yaml "${PACK_DIR}"/

    # kubectl 插件 (kubectl-plugin/ 子目录; install.sh 源在 build/, 打包时并入)
    cp -r "${COMPONENT_DIR}"/kubectl-plugin "${PACK_DIR}"/
    cp "${CUR_DIR}"/install.sh "${PACK_DIR}"/kubectl-plugin/install.sh
    chmod +x "${PACK_DIR}"/kubectl-plugin/install.sh \
        "${PACK_DIR}"/kubectl-plugin/kubectl-ascend_diag \
        "${PACK_DIR}"/kubectl-plugin/kubectl-clusterops

    # 打 zip (交付包); 保留 kubectl-plugin/ 子目录结构 (故不用 -j 全扁平)
    cd "${PACK_DIR}"
    rm -f "${COMPONENT_DIR}/${zip_name}"
    zip "${COMPONENT_DIR}/${zip_name}" \
        agent_core-*.whl \
        node_collector-*.whl \
        agent-core.yaml \
        node-collector.yaml \
        Dockerfile \
        collect_manifest.yaml \
        kubectl-plugin/install.sh \
        kubectl-plugin/kubectl-ascend_diag \
        kubectl-plugin/kubectl-clusterops

    echo
    echo "==> 交付包: ${COMPONENT_DIR}/${zip_name} =="
    unzip -l "${COMPONENT_DIR}/${zip_name}"
}

function main() {
    clean
    check_ascend_fd
    build_whl
    package
}

main
