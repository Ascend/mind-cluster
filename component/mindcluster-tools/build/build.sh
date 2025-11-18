#!/bin/bash

# Perform build mindcluster-tools
# Copyright @ Huawei Technologies CO., Ltd. 2025. All rights reserved
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

CUR_PATH=$(cd "$(dirname "$0")" || exit; pwd)
ROOT_PATH=$(readlink -f "${CUR_PATH}"/..)
CI_PACKAGE_DIR="${ROOT_PATH}"/output/
OUTPUT_DIR="${ROOT_PATH}"/_package_output_py3
BUILD_DIR="${ROOT_PATH}"/build/

function log_base() {
    echo "$(date "+%Y-%m-%d %H:%M:%S") [$1]: $2 ${*:3}"
}

shopt -s expand_aliases
alias log_error='log_base ERROR $LINENO'
alias log_info='log_base INFO $LINENO'
alias log_warn='log_base WARN $LINENO'
alias log_debug='log_base DEBUG $LINENO'

function check_result() {
    ret=$?
    message=$1

    if [ $ret -eq 0 ]; then
       log_info "${message} success."
       return 0
    else
       log_error "${message} failed."
       exit 1
    fi
}

function clean_output() {
    if [ -d "${OUTPUT_DIR}" ]; then
       rm -rf "${OUTPUT_DIR}"
    fi
    check_result "clean output dir"
}

function package() {
    # copy package
    mkdir -p "${OUTPUT_DIR}"

    cp -r "${ROOT_PATH}"/dist/* "${OUTPUT_DIR}"/
    check_result "copy ci package output dir"

    #package
    mkdir -p -m 700 "$CI_PACKAGE_DIR"
    log_info "start build output package"
    cd "${OUTPUT_DIR}"
    cp -r ./* "${CI_PACKAGE_DIR}"
    chmod 400 "${CI_PACKAGE_DIR}"/*
}

function clean() {
    [ -d "${ROOT_PATH}/dist" ] && rm -rf "${ROOT_PATH}/dist"
    rm -rf "${ROOT_PATH}"/*.egg-info
    rm -rf "${BUILD_DIR}"/bdist*
    rm -rf "${BUILD_DIR}/lib"
    check_result "clean"
}

function build_wheel_package() {
    cd "${ROOT_PATH}"
    python3 ./setup.py bdist_wheel --python-tag py3
    check_result "prepare resource"
}

function install_python_requirements() {
  pip3 install -r ${ROOT_PATH}/requirements.txt
}

function main() {
    install_python_requirements
    clean_output
    build_wheel_package
    package
    clean
    clean_output
}

main