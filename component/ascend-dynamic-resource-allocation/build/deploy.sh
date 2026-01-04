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

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
KIND_CLUSTER_NAME="dra-cluster"
KIND_IMAGE="kindest/node:v1.34.0"
KIND_CLUSTER_CONFIG_PATH=${CUR_DIR}/kind-cluster-config.yaml

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

function kind_deploy() {

    kind delete cluster --name ${KIND_CLUSTER_NAME}

    kind create cluster \
    	--name "${KIND_CLUSTER_NAME}" \
    	--image "${KIND_IMAGE}" \
    	--config "${KIND_CLUSTER_CONFIG_PATH}" \
    	--wait 2m

    mkdir ${TOP_DIR}/kind-images

    docker save -o ${TOP_DIR}/kind-images/"driver_images.tar" "ascend-npu-dra-driver:v0.1.0"
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"driver_images.tar"

    ls ${TOP_DIR}/kind-images/ubuntu_images.tar
    if [ $? -ne 0 ]; then
      docker save -o ${TOP_DIR}/kind-images/"ubuntu_images.tar" "ubuntu:22.04"
    fi
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"ubuntu_images.tar"

    ls ${TOP_DIR}/kind-images/cert-manager-controller.tar
    if [ $? -ne 0 ]; then
      docker save -o ${TOP_DIR}/kind-images/"cert-manager-controller.tar" "quay.io/jetstack/cert-manager-controller:v1.16.3"
    fi
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"cert-manager-controller.tar"

    ls ${TOP_DIR}/kind-images/cert-manager-cainjector.tar
    if [ $? -ne 0 ]; then
      docker save -o ${TOP_DIR}/kind-images/"cert-manager-cainjector.tar" "quay.io/jetstack/cert-manager-cainjector:v1.16.3"
    fi
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"cert-manager-cainjector.tar"

    ls ${TOP_DIR}/kind-images/cert-manager-webhook.tar
    if [ $? -ne 0 ]; then
      docker save -o ${TOP_DIR}/kind-images/"cert-manager-webhook.tar" "quay.io/jetstack/cert-manager-webhook:v1.16.3"
    fi
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"cert-manager-webhook.tar"

    ls ${TOP_DIR}/kind-images/cert-manager-startupapicheck.tar
    if [ $? -ne 0 ]; then
      docker save -o ${TOP_DIR}/kind-images/"cert-manager-startupapicheck.tar" "quay.io/jetstack/cert-manager-startupapicheck:v1.16.3"
    fi
    kind load image-archive \
            --name "${KIND_CLUSTER_NAME}" \
            ${TOP_DIR}/kind-images/"cert-manager-startupapicheck.tar"

}

function apply_dra() {
    kubectl create namespace dra-example-driver
    kubectl apply -f ${TOP_DIR}/dra-example-driver-no-webhook.yaml
    kubectl get pods -A
}

function main() {
  go_build
  build_driver
  kind_deploy
  apply_dra
}

main