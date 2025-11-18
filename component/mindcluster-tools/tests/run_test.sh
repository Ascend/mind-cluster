#!/bin/bash
# Perform test mindcluster_tools
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
TOP_DIR=$(realpath "${CUR_PATH}"/..)
echo ${TOP_DIR}
export PYTHONPATH="${TOP_DIR}:${PYTHONPATH}"
PYTHON_PKG=${TOP_DIR}/mindcluster_tools
OUTPUT_DIR="${TOP_DIR}"/test/ut
TEMP_DIR=${TOP_DIR}/tests/ut/test

function install_python_requirements() {
  pip3 install -r ${CUR_PATH}/requirements.txt
}

function build_mock() {
  gcc -std=c11 -fpic -fPIE -shared ${TOP_DIR}/tests/mock/dcmi_mock.c -o ${TOP_DIR}/tests/mock/libdcmi.so
}

function unit_test() {
    export LD_LIBRARY_PATH=${TOP_DIR}/tests/mock:$LD_LIBRARY_PATH
    python3 -m pytest -s --cov=${PYTHON_PKG} --cov-report=html --cov-report=xml --junit-xml=${TEMP_DIR}/test.xml \
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
    rm ${TOP_DIR}/tests/mock/libdcmi.so
}

function execute_test() {
    install_python_requirements
    build_mock
    unit_test
    clean_end
}

execute_test
