#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
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
umask 077
CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/../../..)
export PATH="${GOPATH}/bin/;$PATH"
export GO111MODULE=on
export GONOSUMDB="*"

function execute_test() {
  cd ${TOP_DIR}
  if ! (go test  -mod=mod -gcflags=all=-l -v -race -coverprofile cov.out ${TOP_DIR}/... >./$file_input); then
    echo '****** go test cases error! ******'
    exit 1
  else
    echo ${file_detail_output}
    ${GOPATH}/bin/gocov convert cov.out | ${GOPATH}/bin/gocov-html >${file_detail_output}
    ${GOPATH}/bin/gotestsum --junitfile "${TOP_DIR}"/test/unit-tests.xml "${TOP_DIR}"/...
  fi
}

file_input='testDockerPlugin.txt'
file_detail_output="${TOP_DIR}/test/api.html"

echo "************************************* Start LLT Test *************************************"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi
if [ -f "$file_input" ]; then
  rm -rf $file_input
fi
execute_test

echo "************************************* End   LLT Test *************************************"

exit 0
