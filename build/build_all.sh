#!/bin/bash

# Perform build mindxdl all component
# Copyright @ Huawei Technologies CO., Ltd. 2024-2024. All rights reserved
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
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
GOPATH=$1
ci_config=$2
servicename=$3
volcano_version=$4

#mindx_dl=$(ls -l "$CUR_DIR" |awk '/^d/ {print $NF}')
#for component in $mindx_dl
#do
  echo "Build mindx dl component is " "$servicename"
  case "$servicename" in
    ascend-for-volcano)
      cp -rf "$GOPATH/$ci_config" $GOPATH/src/volcano.sh/volcano/
      ls -la $GOPATH/src/volcano.sh/volcano/
      echo "********$volcano_version*********"
      cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
      dos2unix *.sh && chmod +x *
      ./build.sh $volcano_version
    ;;
    *)
      cp -rf "$GOPATH/$ci_config" $GOPATH/${servicename}/
      ls -la $GOPATH/
      cd $GOPATH/${servicename}/build
      dos2unix *.sh && chmod +x *
      ./build.sh
  esac
#done

