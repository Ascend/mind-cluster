#!/bin/bash
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

OUTPUT_PATH="${ROOT_PATH}/output"
SRC_PATH="${ROOT_PATH}/src"
DT_RESULT_PATH="${OUTPUT_PATH}/DT"
DT_RESULT_COV_DIR="${OUTPUT_PATH}/DT/cov_data"
DT_RESULT_XML_DIR="${OUTPUT_PATH}/DT/xmls"
DT_RESULT_HTMLS_DIR="${OUTPUT_PATH}/DT/htmls"
TOOLKIT_DT_RESULT_COV_DIR="${OUTPUT_PATH}/DT/toolkit_cov_data"

ASCEND_FD_CODE_PATH="${ROOT_PATH}/src/ascend_fd"
ASCEND_FD_DT_REQUIREMENTS_FILE_PATH="${ROOT_PATH}/test/requirements.txt"
ASCEND_FD_REQUIREMENTS_FILE_PATH="${SRC_PATH}/requirements.txt"

TOOLKIT_SRC_PATH="${ROOT_PATH}/toolkit_src"
TOOLKIT_CODE_PATH="${TOOLKIT_SRC_PATH}/ascend_fd_tk"
TOOLKIT_REQUIREMENTS_FILE_PATH="${TOOLKIT_SRC_PATH}/requirements.txt"
ASCEND_CONF_MODEL_DIR="${SRC_PATH}/ascend_fd/configuration/model/"

DECISION_TREE_MODEL_PATH="${ROOT_PATH}/platform/res_model_training/cpu_decision_tree_latest.pkl"
NET_RF_MODEL_PATH="${ROOT_PATH}/platform/net_model_training/net_rf_model_latest.pt"

# setuptools 65.6.3 is incompatible with Python 3.12+ (pkgutil.ImpImporter removed); pin it only for older interpreters.
if python3 -c 'import sys; sys.exit(0 if sys.version_info < (3, 12) else 1)'; then
    pip3 install setuptools==65.6.3
fi

function train_net_model() {
    echo "Begin to train random forest model"

    python3 "${ROOT_PATH}/platform/net_model_training/rf_train.py" -model_path "${NET_RF_MODEL_PATH}"
    if [ ! -d "${ASCEND_CONF_MODEL_DIR}" ]; then
      mkdir "${ASCEND_CONF_MODEL_DIR}"
    fi
    if [ ! -f "${NET_RF_MODEL_PATH}" ]; then
      log_error "network random forest training error"
      exit 1
    fi
    cp "${NET_RF_MODEL_PATH}" "${ASCEND_CONF_MODEL_DIR}"
}

function train_res_model() {
    echo "Begin to train res decision tree model"

    export PYTHONPATH=${ROOT_PATH}/src:$PYTHONPATH
    python3 "${ROOT_PATH}/platform/res_model_training/decision_tree_train.py" -model_path "${DECISION_TREE_MODEL_PATH}"
    if [ ! -d "${ASCEND_CONF_MODEL_DIR}" ]; then
      mkdir "${ASCEND_CONF_MODEL_DIR}"
    fi
    if [ ! -f "${DECISION_TREE_MODEL_PATH}" ]; then
      log_error "cpu decision tree training error"
      exit 1
    fi
    cp "${DECISION_TREE_MODEL_PATH}" "${ASCEND_CONF_MODEL_DIR}"
}

function init_write_build_time() {
    echo "Begin to write_build_time_to_version_info"
    time=$(date "+%Y-%m-%d")
    echo -e "\n${time}" >> "${SRC_PATH}/ascend_fd/Version.info"
}

function build_ascend_fd() {
  echo "build_ascend_fd"
  cd "${ROOT_PATH}"
  train_net_model
  train_res_model
  init_write_build_time

  cd ${SRC_PATH} || exit 3

  python3 setup_linux.py develop -i https://repo.huaweicloud.com/repository/pypi/simple/
}

function run_toolkit_test() {
  echo "Begin to run DT test for ascend_fd_tk"
  cd "${ROOT_PATH}" || exit 3
  pip3 install -r "${TOOLKIT_REQUIREMENTS_FILE_PATH}"
  PYTHONPATH="${TOOLKIT_SRC_PATH}${PYTHONPATH:+:${PYTHONPATH}}" pytest "${TOOLKIT_CODE_PATH}/test" \
  --cov="${TOOLKIT_CODE_PATH}" --cov-branch \
  --junit-xml="${DT_RESULT_XML_DIR}/toolkit.xml" \
  --html="${DT_RESULT_HTMLS_DIR}/toolkit.html" \
  --self-contained-html

  mkdir -p "${TOOLKIT_DT_RESULT_COV_DIR}"
  mv .coverage "${TOOLKIT_DT_RESULT_COV_DIR}"
  cd "${TOOLKIT_DT_RESULT_COV_DIR}" || exit 3
  coverage xml
  coverage html -d coverage_result
  echo "Running DT for ascend_fd_tk over."
}

function run_test() {
  echo "Begin to run DT test for ascend_fd"
  pytest ${ROOT_PATH} \
  --ignore="${TOOLKIT_SRC_PATH}" \
  --cov="${ASCEND_FD_CODE_PATH}" --cov-branch \
  --junit-xml="${DT_RESULT_XML_DIR}/final.xml" \
  --html="${DT_RESULT_HTMLS_DIR}/asecndfd.html" \
  --self-contained-html

  mkdir -p "${DT_RESULT_COV_DIR}"
  mv .coverage "${DT_RESULT_COV_DIR}"
  cd ${DT_RESULT_COV_DIR} || exit 3
  coverage xml
  coverage html -d coverage_result
  echo "Running DT for ascend_fd over."

  run_toolkit_test

  cd ${SRC_PATH} || exit 3
  python3 setup_linux.py develop --uninstall
}

function main() {
  build_ascend_fd
  run_test
  pwd
}

echo "Running DT for ascend_fd now..."
main
echo "All DT for ascend_fd over, now working in dir:"
