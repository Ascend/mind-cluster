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
cd /usr1/mindxdl/component

CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
GOPATH=$1
ci_config=$2

mindx_dl=$(ls -l "$CUR_DIR" |awk '/^d/ {print $NF}')
cd /usr1/mindxdl/build
dos2unix *.sh && chmod +x *

for component in $mindx_dl
do
  { ./build_each.sh $GOPATH $ci_config $component
  }&
done
wait
echo "all component has built"

