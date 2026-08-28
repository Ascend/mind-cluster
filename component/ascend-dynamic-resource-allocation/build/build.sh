#!/bin/bash
# Perform build ascend-dynamic-resource-allocation
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
set -e

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

build_version="v6.0.0"

dra_name="ascend-dra"
build_type=build

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GONOSUMDB="*"
    build_type=ci
fi

if [ "$1" == "edge" ]; then
   build_scene="edge"
fi

function clean() {
    rm -rf "${TOP_DIR}"/output
    mkdir -p "${TOP_DIR}"/output
}

function build_plugin() {
    cd "${TOP_DIR}"
    export CGO_ENABLED=1
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    go build -mod=mod -buildmode=pie -v -gcflags="all=-N -l" -ldflags "-X main.BuildName=${dra_name} \
        -X main.BuildVersion=${build_version}_linux-${os_type} \
        -buildid none" \
        -o ascend-dra  ${TOP_DIR}
    ls "${dra_name}"
    if [ $? -ne 0 ]; then
        echo "fail to find ascend-dra"
        exit 1
    fi
}

function mv_file() {
    mv "${TOP_DIR}/${dra_name}"   "${TOP_DIR}"/output
    cp "${TOP_DIR}/build/Dockerfile"   "${TOP_DIR}"/output
    cp "${TOP_DIR}/build/ascend-dra-driver.yaml"   "${TOP_DIR}"/output
    cp "${TOP_DIR}/build/agreement.txt"   "${TOP_DIR}"/output
}

function change_mod() {
    chmod 400 "$TOP_DIR"/output/*
    chmod 500 "${TOP_DIR}/output/${dra_name}"
}

function main() {
  clean
  build_plugin
  mv_file
  change_mod
}

main $1
