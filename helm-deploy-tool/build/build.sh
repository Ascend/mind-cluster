#!/bin/bash

# Perform  build helm-deploy-tool
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
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v26.1.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  build_version="v"${line#*=}
fi

cd "$TOP_DIR"/build
dos2unix *.sh && chmod +x *

function clear_env() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}/output"
}

function generate_yaml_from_component() {
  ./split_yaml.sh
}

function update_version() {
  REL_NPU_PLUGIN=volcano-npu_${build_version}
  BASE_VER=v1.7.0
  sed -i "s/name: volcano-npu_v6.0.RC1_linux-x86_64/name: ${REL_NPU_PLUGIN}/" "${TOP_DIR}/app/charts/ascend-for-volcano/yamls/${BASE_VER}"/ConfigMap-*.yaml
  BASE_VER=v1.9.0
  sed -i "s/name: volcano-npu_v6.0.RC1_linux-x86_64/name: ${REL_NPU_PLUGIN}/" "${TOP_DIR}/app/charts/ascend-for-volcano/yamls/${BASE_VER}"/ConfigMap-*.yaml
}

function replace_yaml_value() {
  python3 ${TOP_DIR}/build/replace_yaml_values.py -v ${build_version}
}

function helm_package() {
  helm package "${TOP_DIR}"/app
  helm package "${TOP_DIR}"/app-crds
}

function mv_file() {
  cp "${TOP_DIR}"/build/add_helm_meta.sh "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/*.tgz "${TOP_DIR}"/output
}

function change_mod() {
    chmod 400 "${TOP_DIR}"/output/*
}

function main() {
  clear_env
  generate_yaml_from_component
  update_version
  replace_yaml_value
  helm_package
  mv_file
  change_mod
}

main
