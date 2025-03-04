#!/bin/bash

# Perform build taskd
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
VER_FILE="${ROOT_PATH}"/service_config.ini
build_version="v7.0.RC1"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
  echo "build version in service_config.ini:  ${build_version}"
fi
export BUILD_VERSION=${build_version}
echo "build version is ${BUILD_VERSION}"

bash "$BUILD_DIR"/build_backend.sh

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
    check_result "clean"
}

function build_wheel_package() {
    cd "${ROOT_PATH}"
    if [[ "$(uname -m)" == "x86_64" ]]; then
      python3 ./setup.py bdist_wheel --plat-name linux_x86_64 --python-tag py3
    else
      python3 ./setup.py bdist_wheel --plat-name linux_aarch64 --python-tag py3
    fi
    check_result "prepare resource"
}

function main() {
    clean_output
    build_wheel_package
    package
    clean
    clean_output
}

main