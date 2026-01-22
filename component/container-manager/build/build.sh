#!/bin/bash
# Perform build container-manager
# Copyright @ Huawei Technologies CO., Ltd. 2025. All rights reserved
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
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v7.3.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v7.3.0
  build_version="v"${line#*=}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

output_name="container-manager"
os_type=$(arch)

function clean() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
    cd "${TOP_DIR}"
    export CGO_ENABLED=1
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    go build -mod=mod -buildmode=pie -ldflags "-X main.BuildName=${output_name} \
            -X main.BuildVersion=${build_version}_linux-${os_type} \
            -buildid none     \
            -s   \
            -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
            -o "${output_name}"  \
            -trimpath
    chmod 0500 "${output_name}"
    if [ $? -ne 0 ]; then
        echo "failed to find component container-manager"
        exit 1
    fi
}

function mv_file() {
    mv "${TOP_DIR}/${output_name}"   "${TOP_DIR}"/output
}

function main() {
  clean
  build
  mv_file
}

main
