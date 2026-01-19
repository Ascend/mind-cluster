#!/bin/bash

# Perform build mind-cluster each component
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


GOPATH=$1
config=$2
servicename=$3

function build_volcano() {
  cp -rf "$GOPATH/$config" $GOPATH/src/volcano.sh/volcano/
  ls -la $GOPATH/src/volcano.sh/volcano/
  echo "********$1*********"
  cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
  dos2unix *.sh && chmod +x *
  ./build.sh $1
}

function build_other() {
  cp -rf "$GOPATH/$config" $GOPATH/${1}/
  ls -la $GOPATH/
  cd $GOPATH/${1}/build
  dos2unix *.sh && chmod +x *
  ./build.sh
}

echo "Build mindx dl component is " "$servicename"
case "$servicename" in
  ascend-for-volcano)
    echo "***************start complie volcano 1.9***********************"
    mkdir -p ${GOPATH}/src/volcano.sh && cp -rf /opt/buildtools/volcano_opensource/volcano_1.9/volcano ${GOPATH}/src/volcano.sh/
    ls -la ./ &&  cp -rf ${GOPATH}/${servicename} ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/
    cd ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ${servicename} ascend-volcano-plugin
    build_volcano v1.9.0
    mkdir -p ${GOPATH}/output/volcano-v1.9.0 && cp -rf ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${GOPATH}/output/volcano-v1.9.0/
    ls -la ${GOPATH}/output/volcano-v1.9.0/
    rm -rf ${GOPATH}/src/volcano.sh/volcano
    echo "***************start complie volcano 1.7***********************"
    cp -rf /opt/buildtools/volcano_opensource/volcano_1.7/volcano ${GOPATH}/src/volcano.sh/
    ls -la ./ &&  cp -rf ${GOPATH}/${servicename} ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/
    cd ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ && mv ${servicename} ascend-volcano-plugin
    build_volcano v1.7.0
    mkdir -p ${GOPATH}/output/volcano-v1.7.0 && cp -rf ${GOPATH}/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output/* ${GOPATH}/output/volcano-v1.7.0/
    ls -la ${GOPATH}/output/volcano-v1.7.0/
    rm -rf ${GOPATH}/src/volcano.sh/volcano

  ;;
  ascend-docker-runtime)
    cd ${GOPATH}/${servicename}/opensource && tar -zxvf makeself/makeself-2.4.2.tar.gz
    build_other ${servicename}

  ;;
  ascend-operator)
    build_other ${servicename}

  ;;
  *)
    build_other ${servicename}

esac





