#!/bin/bash
# Perform build noded
# Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE=on
version_file="${TOP_DIR}"/service_config.ini
build_version="v6.0.0"
if  [ -f "$version_file" ]; then
  line=$(sed -n '1p' "$version_file" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

sed -i "s/noded:.*/noded:${build_version}/" "${TOP_DIR}"/build/noded.yaml
sed -i "s/noded:.*/noded:${build_version}/" "${TOP_DIR}"/build/noded-dpc.yaml

OUTPUT_NAME="noded"
DOCKER_FILE_NAME="Dockerfile"
NODE_CONFIG_FILE_NAME="NodeDConfiguration.json"
function clean() {
    rm -rf "${TOP_DIR}"/output
    mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}"
  export CGO_ENABLED=1
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now
    -X main.BuildVersion=${build_version}_linux-${arch} \
    -X main.BuildName=${OUTPUT_NAME}" \
    -o ${OUTPUT_NAME}

  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
      echo "fail to find ${OUTPUT_NAME}"
      exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}/${OUTPUT_NAME}" "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/noded.yaml "${TOP_DIR}"/output/noded-"${build_version}".yaml
  cp "${TOP_DIR}"/build/noded-dpc.yaml "${TOP_DIR}"/output/noded-dpc-"${build_version}".yaml
  cp "${TOP_DIR}"/build/pingmesh-config.yaml "${TOP_DIR}"/output/pingmesh-config.yaml
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/${NODE_CONFIG_FILE_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/fdConfig.yaml "${TOP_DIR}"/output
  chmod 400 "${TOP_DIR}"/output/*
  chmod 500 "${TOP_DIR}"/output/${OUTPUT_NAME}
}

function main() {
  clean
  build
  mv_file
}

main
