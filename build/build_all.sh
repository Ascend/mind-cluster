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

cp -rf "$TOP_DIR"/component/* ${GOPATH}/
if [[ ! -d /opt/buildtools/volcano_opensource ]]; then
    mkdir -p /opt/buildtools/volcano_opensource/volcano_1.9/
    cd /opt/buildtools/volcano_opensource/volcano_1.9/
    git clone -b release-1.9 https://github.com/volcano-sh/volcano.git
    mkdir -p /opt/buildtools/volcano_opensource/volcano_1.7/
    cd /opt/buildtools/volcano_opensource/volcano_1.7/
    git clone -b release-1.7 https://github.com/volcano-sh/volcano.git
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
mind_cluster=$(ls -l "$CUR_DIR" |awk '/^d/ {print $NF}')
cd "$TOP_DIR"/build
cp -rf "$TOP_DIR"/build/service_config.ini $GOPATH/service_config.ini
dos2unix *.sh && chmod +x *

for component in $mind_cluster
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

for component in $mind_cluster
do
  {
    if [[ $component = "ascend-common" ]]; then
      continue
    fi
    if [[ $component = "ascend-for-volcano" ]]; then
      cd "$TOP_DIR"/component/"$component"
      rm -rf ./output
      mv "$GOPATH"/output/ ./
      zip -r ./output/Ascend-mindxdl-volcano_linux.zip ./output/*
      continue
    fi
    cd "$TOP_DIR"/component/"$component"
    rm -rf ./output
    mv "$GOPATH"/"$component"/output/ ./
    zip -r ./output/Ascend-mindxdl-"$component"_linux.zip ./output/*
  }
done
