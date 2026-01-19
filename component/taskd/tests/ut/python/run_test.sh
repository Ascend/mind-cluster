#!/bin/bash

# Perform test taskd-python
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
CUR_PATH=$(cd "$(dirname $0)" || exit; pwd)
TEST_MODULE=("framework")
TOP_DIR=$(realpath "${CUR_PATH}"/../../..)
export PYTHONPATH="${TOP_DIR}:${CUR_PATH}:${PYTHONPATH}"
PYTHON_PKG=${TOP_DIR}/taskd/python
OUTPUT_DIR="${TOP_DIR}"/test/ut/python
TEMP_DIR=${TOP_DIR}/tests/ut/python/test

function unit_test() {
    python3 -m pytest --cov=${PYTHON_PKG} --cov-report=html --cov-report=xml \
    --cov-config=${CUR_PATH}/.coveragerc --junit-xml=${TEMP_DIR}/test.xml \
    --html=${TEMP_DIR}/api.html --self-contained-html --durations=5 -vv --cov-branch
    RET=$?
    if [ ${RET} -ne 0 ]; then
         exit ${RET}
    fi
    python3 -m coverage xml -i
    mv coverage.xml .coverage ${TEMP_DIR}/
    mv htmlcov ${TEMP_DIR}/
}


function clean_before() {
    if [ -d "$TEMP_DIR" ]; then
      rm -rf $TEMP_DIR
    fi
}

function clean_end() {
    if [ -d "${OUTPUT_DIR}" ]; then
       rm -rf "${OUTPUT_DIR}"
    fi
    mkdir -p "${OUTPUT_DIR}"
    mv "${TEMP_DIR}"/* "${OUTPUT_DIR}"/
    rm -rf "${TEMP_DIR}"
}

function execute_test() {
    pip3 install -r "${CUR_PATH}"/requirements.txt
    unit_test
    clean_end
}

execute_test