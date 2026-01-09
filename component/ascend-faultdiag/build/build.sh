#!/bin/bash
# -*- coding:utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
# ==============================================================================

set -e

CUR_PATH=$(cd "$(dirname "$0")" || exit; pwd)
ROOT_PATH=$(readlink -f "$CUR_PATH"/..)
SRC_PATH="${ROOT_PATH}/src"
OUTPUT_PATH="${ROOT_PATH}/output/"
ASCEND_CONF_MODEL_DIR="${SRC_PATH}/ascend_fd/configuration/model/"

DECISION_TREE_LATEST_MODEL_PATH="${ROOT_PATH}/platform/res_model_training/cpu_decision_tree_latest.pkl"
NET_RF_LATEST_MODEL_PATH="${ROOT_PATH}/platform/net_model_training/net_rf_model_latest.pt"
DECISION_TREE_OLD_MODEL_PATH="${ROOT_PATH}/platform/res_model_training/cpu_decision_tree_102.pkl"
NET_RF_OLD_MODEL_PATH="${ROOT_PATH}/platform/net_model_training/net_rf_model_102.pt"


function log_base() {
    echo "$(date "+%Y-%m-%d %H:%M:%S") [$1]: $2 ${*:3}"
}

shopt -s expand_aliases
alias log_error='log_base ERROR $LINENO'
alias log_info='log_base INFO $LINENO'
alias log_warn='log_base WARN $LINENO'
alias log_debug='log_base DEBUG $LINENO'

function check_version() {
    BUILD_VERSION="7.3.0"
    local version_file="${ROOT_PATH}/service_config.ini"
    if  [ -f "$version_file" ]; then
      line=$(sed -n '1p' "$version_file" 2>&1)
      BUILD_VERSION="v"${line#*=}
    fi
}

function check_result() {
    ret=$?
    message=$1

    if [ $ret -eq 0 ]; then
      log_info "$message success."
      return 0
    else
      log_error "$message failed."
      exit 1
    fi
}

function clear() {
    log_info "Begin to clear package files"
    rm -rf "${SRC_PATH}/ascend_fd.egg-info"
    rm -rf "${SRC_PATH}/dist"
    rm -rf "${SRC_PATH}/build"
}

function init_kg_engine_expr_parser() {
    log_info "Begin to initialize expr_parser"
    python3 "${ROOT_PATH}/platform/expr_parser_initializing/generate_parser_out.py"
}

function train_net_model() {
    log_info "Begin to train net random forest model in $1"

    "$1" "${ROOT_PATH}/platform/net_model_training/rf_train.py" -model_path "$2"
    if [ ! -d "${ASCEND_CONF_MODEL_DIR}" ]; then
      mkdir "${ASCEND_CONF_MODEL_DIR}"
    fi
    if [ ! -f "$2" ]; then
      log_error "network random forest training error in $1"
      exit 1
    fi
    cp "$2" "${ASCEND_CONF_MODEL_DIR}"
}

function train_res_model() {
    log_info "Begin to train res decision tree model in $1"

    "$1" "${ROOT_PATH}/platform/res_model_training/decision_tree_train.py" -model_path "$2"
    if [ ! -d "${ASCEND_CONF_MODEL_DIR}" ]; then
      mkdir "${ASCEND_CONF_MODEL_DIR}"
    fi
    if [ ! -f "$2" ]; then
      log_error "cpu decision tree training error"
      exit 1
    fi
    cp "$2" "${ASCEND_CONF_MODEL_DIR}"
}

function compile_build() {
    log_info "Begin to build ascend_faultdiag package"
    cd ${SRC_PATH}
    local lib_path="${ROOT_PATH}/src/ascend_fd/lib"
    echo "Downloading Fault Diagnosis library..."
    if [[ "$(uname -m)" == "x86_64" ]]; then
      PLAT_FORM="linux_x86_64"
      mkdir -p "$lib_path" && wget -O "${lib_path}/libfaultdiag.so" https://mindcluster.obs.cn-north-4.myhuaweicloud.com/ascend-repo/libfaultdiag_x86_64.so --no-check-certificate
    else
      PLAT_FORM="linux_aarch64"
      mkdir -p "$lib_path" && wget -O "${lib_path}/libfaultdiag.so" https://mindcluster.obs.cn-north-4.myhuaweicloud.com/ascend-repo/libfaultdiag_aarch64.so --no-check-certificate
    fi
    chmod 640 ${ROOT_PATH}/src/ascend_fd/configuration/*.json
    python3 ./setup_linux.py --mode zh --version $BUILD_VERSION bdist_wheel --plat-name $PLAT_FORM
    check_result "build ascend_faultdiag package"
    log_info "Begin to mv ascend_faultdiag.whl to ${OUTPUT_PATH}"
    cp -rf "${SRC_PATH}"/dist/ascend_faultdiag*.whl "${OUTPUT_PATH}"
    log_info "Begin to mv other files to ${OUTPUT_PATH}"
    python3 "${ROOT_PATH}/platform/international_pkg_config.py" --path "${SRC_PATH}" --old ascend_fd --new alan_fd
    mv "${ROOT_PATH}/src/ascend_fd" "${ROOT_PATH}/src/alan_fd"
    cd ${SRC_PATH}
    python3 ./setup_linux.py --mode en --version $BUILD_VERSION bdist_wheel --plat-name $PLAT_FORM
    check_result "build alan_faultdiag package"
    log_info "Begin to mv alan_faultdiag.whl to ${OUTPUT_PATH}"
    cp -rf "${SRC_PATH}"/dist/alan_faultdiag*.whl "${OUTPUT_PATH}"
}

function main() {
    export PYTHONPATH=${ROOT_PATH}/src:$PYTHONPATH
    # Create the final output path
    mkdir -p ${OUTPUT_PATH}
    train_net_model python3 "${NET_RF_LATEST_MODEL_PATH}"
    train_res_model python3 "${DECISION_TREE_LATEST_MODEL_PATH}"
    init_kg_engine_expr_parser
    check_version
    compile_build
    chmod 640 ${OUTPUT_PATH}/*.whl
    clear
}

echo "begin to build ascend_fd"
main;ret=$?
echo "finish build ascend_fd, check_result is $ret"