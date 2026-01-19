#!/bin/bash

# Perform test taskd-ut
# Copyright(C) Huawei Technologies Co.,Ltd. 2025. All rights reserved.
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

CUR_PATH=$(cd "$(dirname $0)";pwd )
TEST_MODULE=("python" "go")
TOP_DIR=$(realpath "${CUR_PATH}"/../..)
OUTPUT_DIR="${TOP_DIR}"/test/ut/

function clean_before() {
    if [ -d "$OUTPUT_DIR" ]; then
      rm -rf $OUTPUT_DIR
    fi
}

function run_test_all() {
    for module in "${TEST_MODULE[@]}"; do
       cd ${CUR_PATH}/$module
       bash run_test.sh
    done
}

function run_test_module() {
    found_module=false
    test_module=$1
    for module in "${TEST_MODULE[@]}"; do
        if [ "$module" = "${test_module}" ]; then
            found_module=true
            break
        fi
    done
    if $found_module; then
       cd ${CUR_PATH}/$test_module
       bash run_test.sh
    else
      echo "wrong module name!"
    fi
}

function run_test() {
    clean_before
    mkdir -p ${OUTPUT_DIR}
    if [ $# -eq 0 ]; then
        run_test_all
    else
        run_test_module
    fi
}

run_test
