#!/bin/bash

# Perform  split yaml files and store to app and app-crds directory
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

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
BASE_DIR=$(cd "${SCRIPT_DIR}/.." && pwd)
COMPONENT_BASE_DIR=$(cd "${BASE_DIR}/../component" && pwd)

CRDS_DIR="${BASE_DIR}/app-crds/charts"
APP_DIR="${BASE_DIR}/app/charts"

process_resource() {
    local doc=$1
    local component_name=$2
    local variant=$3

    local kind
    local name
    kind=$(echo "$doc" | yq eval '.kind' -)
    name=$(echo "$doc" | yq eval '.metadata.name' -)

    if [[ -z "${kind}" || "${kind}" == "null" || -z "${name}" || "${name}" == "null" ]]; then
        echo "Warning: skipping resource with kind='${kind}', name='${name}'"
        return
    fi

    local filename="${kind}-${name}.yaml"
    local target_dir

    if [[ "${kind}" == "CustomResourceDefinition" ]]; then
        target_dir="${CRDS_DIR}/${component_name}-crds/yamls"
        if [[ -n "${variant}" ]]; then
            target_dir="${target_dir}/${variant}"
        fi
    else
        target_dir="${APP_DIR}/${component_name}/yamls"
        if [[ -n "${variant}" ]]; then
            target_dir="${target_dir}/${variant}"
        fi
    fi

    mkdir -p "${target_dir}"
    local filepath="${target_dir}/${filename}"
    echo "$doc" > "${filepath}"
    echo "Created: ${filepath}"
}

process_component() {
    local source_yaml=$1
    local component_name=$2
    local variant=$3

    if [[ ! -f "${source_yaml}" ]]; then
        echo "Warning: source YAML file not found: ${source_yaml}, skipping"
        return
    fi

    if [[ -n "${variant}" ]]; then
        echo "========== Processing: ${source_yaml} -> ${component_name} (variant: ${variant}) =========="
    else
        echo "========== Processing: ${source_yaml} -> ${component_name} =========="
    fi

    local i=0
    local doc
    while true; do
        doc=$(yq eval "select(documentIndex == $i)" "$source_yaml")
        if [[ -z "$doc" || "$doc" == "null" ]]; then
            break
        fi
        process_resource "$doc" "$component_name" "$variant"
        ((i++))
    done

    if [[ $i -eq 0 ]]; then
        echo "Warning: no documents found in ${source_yaml}"
    fi
}

declare -A COMPONENT_PATTERNS=(
    ["infer-operator"]="infer-operator*.yaml"
    ["ascend-device-plugin"]="ascendplugin-*.yaml"
    ["clusterd"]="clusterd*.yaml"
    ["noded"]="noded*.yaml"
    ["ascend-operator"]="ascend-operator*.yaml"
    ["npu-exporter"]="npu-exporter*.yaml"
    ["ascend-for-volcano"]="volcano*.yaml"
)

for component in "${!COMPONENT_PATTERNS[@]}"; do
    pattern="${COMPONENT_PATTERNS[$component]}"
    build_dir="${COMPONENT_BASE_DIR}/${component}/build"

    shopt -s nullglob
    files=("${build_dir}"/${pattern})
    shopt -u nullglob

    for yaml_file in "${files[@]}"; do
        base=$(basename "${yaml_file}" .yaml)
        variant="default"
        if [[ "${component}" == "ascend-device-plugin" ]]; then
            variant="${base#ascendplugin-}"
        elif [[ "${component}" == "ascend-for-volcano" ]]; then
            variant="${base#volcano-}"
        else
            if [[ "${component}" == "${base}" ]]; then
                variant="default"
            else
              variant="${base#${component}-}"
            fi
        fi
        process_component "${yaml_file}" "${component}" "${variant}"
    done
done

echo "========== Done =========="
