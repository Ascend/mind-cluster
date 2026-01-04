#!/bin/bash
# Deploy ascend-dynamic-resource-allocation
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

# To deploy, we need:
# kubernetes v1.34+ with containerd runtime
# image: ubuntu22.04
# go v1.24+


CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
SLEEP_TIME=2

function go_build() {
    ${CUR_DIR}/build.sh
    if [ $? -ne 0 ]; then
        echo "go build fail"
        exit 1
    fi
}

function build_driver() {
    docker build -f ${CUR_DIR}/Dockerfile -t ascend-npu-dra-driver:v0.1.0 ${TOP_DIR}
    if [ $? -ne 0 ]; then
        echo "images build fail"
        exit 1
    fi
}


function apply_dra() {
    kubectl create namespace ascend-npu-dra-driver
    docker save -o plugin.tar ascend-npu-dra-driver:v0.1.0
    nerdctl -n k8s.io load -i plugin.tar
    rm plugin.tar
    kubectl delete -f ${TOP_DIR}/dra-example-driver-no-webhook.yaml
    kubectl apply -f ${TOP_DIR}/dra-example-driver-no-webhook.yaml
    sleep $SLEEP_TIME
    kubectl get pods -A
}

function main() {
  go_build
  build_driver
  apply_dra
}

main