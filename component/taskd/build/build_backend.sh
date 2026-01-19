#!/bin/bash

# Perform build taskd_backend
# Copyright @ Huawei Technologies CO., Ltd. 2025. All rights reserved
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
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version=${BUILD_VERSION}
output_name="libtaskd.so"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  # cut the chars after '=' and add char 'v', the final example is v3.0.0
  # change for pep 440 specification
  temp=${line#*=}
  build_version="v${temp//.SPC/+SPC}"
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"
os_type=$(arch)

function clean() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}/taskd/go"
  go mod tidy
  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=c-shared -ldflags "-X main.BuildName=${output_name} \
              -X main.BuildScene=${build_scene} \
              -X main.BuildVersion=${build_version}_linux-${os_type} \
              -buildid none \
              -s \
              -linkmode=external \
              -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
              -o "${output_name}"  \
              -trimpath
  ls "${output_name}"
  if [ $? -ne 0 ]; then
      echo "fail to find taskd"
      exit 1
  fi
}

function mv_file() {
  rm -rf "${TOP_DIR}"/taskd/python/cython_api/libs
  mkdir -p "${TOP_DIR}"/taskd/python/cython_api/libs
  mv "${output_name}" "${TOP_DIR}"/taskd/python/cython_api/libs
  cd "${TOP_DIR}"
}

function main() {
  clean
  build
  mv_file
}

main
