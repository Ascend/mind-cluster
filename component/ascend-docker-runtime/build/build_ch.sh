#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
# Description: ascend-docker-runtime build script
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================


set -e

Ascend="Alan"
ascend="alan"

ROOT=$(cd $(dirname $0); pwd)/..

modify_files=(${ROOT}/build/scripts/run_main.sh ${ROOT}/build/scripts/uninstall.sh ${ROOT}/build/makeself-header/makeself-header.sh ${ROOT}/build/build.sh)
for cur_file in "${modify_files[@]}"
do
  sed -i "s/RT_LOWER_CASE=\"ascend-docker-runtime\"/RT_LOWER_CASE=\"${ascend}-docker-runtime\"/g" ${cur_file}
  sed -i "s/RT_FIRST_CASE=\"Ascend-docker-runtime\"/RT_FIRST_CASE=\"${Ascend}-docker-runtime\"/g" ${cur_file}
done

sed -i "s/ascend-docker-runtime/${ascend}-docker-runtime/g" ${ROOT}/build/scripts/help.info

sed -i "s#/var/log/ascend-docker-runtime/#/var/log/${ascend}-docker-runtime/#g" ${ROOT}/cli/src/logger.c

chmod +x build.sh && dos2unix build.sh
source build.sh