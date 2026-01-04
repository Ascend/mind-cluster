#!/bin/bash
# Perform build ascend-dynamic-resource-allocation
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

build_version="v0.1.0"
os_type=$(uname -m)
dra_name="dra-kubelet-plugin"

function clean() {
    rm -rf "${TOP_DIR}"/output
    mkdir -p "${TOP_DIR}"/output
}

function build_plugin() {
    cd "${TOP_DIR}"
    export CGO_ENABLED=0
    go build -mod=mod -buildmode=pie -v -ldflags "-X main.BuildName=${dra_name} \
        -X main.BuildVersion=${build_version}_linux-${os_type} \
        -buildid none     \
        -s   \
        -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
        -o $dra_name  ${TOP_DIR}/pkg/cmd/dra-npu-kubeletplugin
    ls "${dra_name}"
    if [ $? -ne 0 ]; then
        echo "fail to find dra-kubelet-plugin"
        exit 1
    fi
}

function mv_file() {
    mv "${TOP_DIR}/${dra_name}"   "${TOP_DIR}"/output
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
