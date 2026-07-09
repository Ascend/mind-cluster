#!/bin/bash

# Perform build mind-cluster all component
# Copyright @ Huawei Technologies CO., Ltd. 2024-2025. All rights reserved
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
GOPATH=$1
NEW_GOPATH="/usr1/gopath"

if [ -z "$GOPATH" ]; then
    export GOPATH="$NEW_GOPATH"
    rm -rf "$NEW_GOPATH"
    mkdir -p "$NEW_GOPATH"
    echo "GOPATH has been set to $GOPATH"
fi

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

# Read version number
VER_FILE="${CUR_DIR}"/service_config.ini
version="v26.1.0"
if [ -f "$VER_FILE" ]; then
    line=$(sed -n '1p' "$VER_FILE" 2>&1)
    version="v"${line#*=}
fi
ostype=$(arch 2>&1)

# Packaging function: Generates zip/run packages
package_component() {
    local serviceName=$1
    local component_dir=$2
    local servicepackName=$3
    local is_run=$4  # if .run type

    cd "${component_dir}"/output/

    local package_name
    if [ "${is_run}" == "true" ]; then
        # ascend-docker-runtime package to .run
        package_name=Ascend-docker-runtime_${version}_linux-${ostype}.run
    else
        package_name=Ascend-mindxdl-${servicepackName}_${version}_linux-${ostype}.zip
        zip -r "${package_name}" ./*
    fi

    echo "Package created: ${package_name}"
}

cp -rf "$TOP_DIR"/component/* ${GOPATH}/
if [[ ! -d /opt/buildtools/volcano_opensource ]]; then
    mkdir -p /opt/buildtools/volcano_opensource/volcano_1.7/
    mkdir -p /opt/buildtools/volcano_opensource/volcano_1.9/
    mkdir -p /opt/buildtools/volcano_opensource/volcano_1.12/
fi

if [[ ! -d /opt/buildtools/volcano_opensource/volcano_1.7/volcano ]]; then
    cd /opt/buildtools/volcano_opensource/volcano_1.7
    git clone -b release-1.7 https://github.com/volcano-sh/volcano.git
fi

if [[ ! -d /opt/buildtools/volcano_opensource/volcano_1.9/volcano ]]; then
    cd /opt/buildtools/volcano_opensource/volcano_1.9
    git clone -b release-1.9 https://github.com/volcano-sh/volcano.git
fi

if [[ ! -d /opt/buildtools/volcano_opensource/volcano_1.12/volcano ]]; then
    cd /opt/buildtools/volcano_opensource/volcano_1.12
    git clone -b release-1.12 https://github.com/volcano-sh/volcano.git
fi

if [[ ! -d ${GOPATH}/ascend-docker-runtime/platform/libboundscheck ]]; then
    mkdir -p ${GOPATH}/ascend-docker-runtime/platform
    cd ${GOPATH}/ascend-docker-runtime/platform
    git clone -b v1.1.10 https://gitee.com/openeuler/libboundscheck.git
fi

if [[ ! -d ${GOPATH}/ascend-docker-runtime/opensource/makeself ]]; then
    mkdir -p ${GOPATH}/ascend-docker-runtime/opensource
    cd ${GOPATH}/ascend-docker-runtime/opensource
    git clone -b openEuler-22.03-LTS https://gitee.com/src-openeuler/makeself.git
    tar -zxvf makeself/makeself-2.4.2.tar.gz
fi

cd "$TOP_DIR"/component
CUR_DIR=$(dirname $(readlink -f $0))
mind_cluster=(
    "ascend-device-plugin"
    "ascend-docker-runtime"
    "ascend-for-volcano"
    "ascend-operator"
    "ascend-faultdiag"
    "clusterd"
    "container-manager"
    "infer-operator"
    "k8s-rdma-shared-dev-plugin"
    "mindio"
    "noded"
    "npu-exporter"
    "taskd"
)
cd "$TOP_DIR"/build
cp -rf "$TOP_DIR"/build/service_config.ini $GOPATH/service_config.ini
dos2unix *.sh && chmod +x *

for component in "${mind_cluster[@]}"
do
  {
    if [[ $component = "ascend-common" ]]; then
      continue
    fi
    ./build_each.sh $GOPATH service_config.ini $component
  }
done
wait
echo "all component has built"

for component in "${mind_cluster[@]}"
do
  {
    if [[ $component = "ascend-common" ]]; then
      continue
    fi

    cd "$TOP_DIR"/component/"$component"
    rm -rf ./output

    # The output of ascend-for-volcano is in GOPATH/output/, and the others are in GOPATH/component/output/.
    if [[ $component = "ascend-for-volcano" ]]; then
      mv "$GOPATH"/output/ ./
    else
      mv "$GOPATH"/"$component"/output/ ./
    fi

    # Select the packaging method based on the component type.
    case "$component" in
      ascend-for-volcano)
        package_component "$component" "$TOP_DIR/component/$component" "volcano" "false"
        ;;
      ascend-device-plugin)
        package_component "$component" "$TOP_DIR/component/$component" "device-plugin" "false"
        ;;
      ascend-docker-runtime)
        package_component "$component" "$TOP_DIR/component/$component" "docker-runtime" "true"
        ;;
      ascend-faultdiag)
        package_component "$component" "$TOP_DIR/component/$component" "faultdiag" "false"
        ;;
      npu-exporter)
        package_component "$component" "$TOP_DIR/component/$component" "npu-exporter" "false"
        ;;
      noded)
        package_component "$component" "$TOP_DIR/component/$component" "noded" "false"
        ;;
      *)
        package_component "$component" "$TOP_DIR/component/$component" "$component" "false"
        ;;
    esac
  }
done

cp -rf "$TOP_DIR"/build/service_config.ini "$TOP_DIR"/helm-deploy-tool/
cd "$TOP_DIR"/helm-deploy-tool/build/
dos2unix *.sh && chmod +x *
./build.sh
echo "helm deploy tool has built"

# package helm-deploy-tool
cd "$TOP_DIR"/helm-deploy-tool/output/
zip -r Ascend-helm-deploy-tool_${version}_linux.zip ./*
echo "Package created: Ascend-helm-deploy-tool_${version}_linux.zip"
