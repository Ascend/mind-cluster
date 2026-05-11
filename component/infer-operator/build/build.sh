#!/bin/bash

# Perform  build infer-operator
# Copyright @ Huawei Technologies CO., Ltd. 2026. All rights reserved
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v6.0.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

OUTPUT_NAME="infer-operator"
sed -i "s/image: infer-operator:.*/image: infer-operator:${build_version}/" "${TOP_DIR}"/build/${OUTPUT_NAME}.yaml
sed -i "s/Infer Operator Version .*/Infer Operator Version ${build_version}/" "${TOP_DIR}"/build/agreement.txt

DOCKER_FILE_NAME="Dockerfile"

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
  go build -mod=mod -buildmode=pie  -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now  -X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version}_linux-${arch}" \
            -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find infer-operator"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/${OUTPUT_NAME}" "${TOP_DIR}/output"
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/Dockerfile.openeuler "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/agreement.txt "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/infer-operator.yaml "${TOP_DIR}"/output/infer-operator-"${build_version}".yaml
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