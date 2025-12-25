#!/bin/bash
# Perform  build rl-operator
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v6.0.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

OUTPUT_NAME="rl-operator"
sed -i "s/rl-operator:.*/rl-operator:${build_version}/" "${TOP_DIR}"/build/${OUTPUT_NAME}.yaml

DOCKER_FILE_NAME="Dockerfile"
CRD_DIR_NAME="crds"

function clear_env() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}/output"
}

function build() {
  cd "${TOP_DIR}"
  export CGO_ENABLED=0
  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go mod tidy
  CGO_ENABLED=1
  go build -mod=mod -buildmode=pie  -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now  -X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version}_linux-${arch}" \
            -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find rl-operator"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/${OUTPUT_NAME}" "${TOP_DIR}/output"
  cp "${TOP_DIR}"/build/rl-operator.yaml "${TOP_DIR}"/output/rl-operator-"${build_version}".yaml
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
  cp -r "${TOP_DIR}"/build/${CRD_DIR_NAME} "${TOP_DIR}"/output
}

function change_mod() {
  chmod 500 "${TOP_DIR}/output/${OUTPUT_NAME}"
  # set YAML and Dockerfile to readonly
  chmod 400 "${TOP_DIR}"/output/*.yaml 2>/dev/null || true
  chmod 400 "${TOP_DIR}"/output/Dockerfile 2>/dev/null || true
  # set permissions of crds
  if [ -d "${TOP_DIR}/output/crds" ]; then
    find "${TOP_DIR}/output/crds" -type f -exec chmod 400 {} \;
    find "${TOP_DIR}/output/crds" -type d -exec chmod 500 {} \;
  fi
}

function main() {
  clear_env
  build
  mv_file
  change_mod
}

main
