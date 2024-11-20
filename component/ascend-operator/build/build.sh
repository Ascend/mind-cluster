#!/bin/bash
# Perform  build ascend-operator
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v5.0.RC1"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"
if [ "${arch:0:5}" = 'aarch' ]; then
  arch=arm64
else
  arch=amd64
fi

OUTPUT_NAME="ascend-operator"
sed -i "s/ascend-operator:.*/ascend-operator:${build_version}/" "${TOP_DIR}"/build/${OUTPUT_NAME}.yaml

DOCKER_FILE_NAME="Dockerfile"
docker_zip_name="ascend-operator-${build_version}-${arch}.tar.gz"
docker_images_name="ascend-operator:${build_version}"

function clear_env() {
  rm -rf "${TOP_DIR}"/output/*
  mkdir -p "${TOP_DIR}/output"
}

function build() {
  cd "${TOP_DIR}"

  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie  -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now  -X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version}" \
            -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find ascend-operator"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/${OUTPUT_NAME}" "${TOP_DIR}/output"
  cp "${TOP_DIR}"/build/ascend-operator.yaml "${TOP_DIR}"/output/ascend-operator-"${build_version}".yaml
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
}

function change_mod() {
  chmod 400 "${TOP_DIR}"/output/*
  chmod 500 "${TOP_DIR}/output/${OUTPUT_NAME}"
}



function main() {
  clear_env
  build
  mv_file
  change_mod
}

main
