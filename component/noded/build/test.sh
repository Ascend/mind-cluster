#!/bin/bash
# Perform test noded
# Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

# execute go test and echo result to report files
function execute_test() {
  if ! (go test -mod=mod -gcflags=all=-l -v -race -coverprofile cov.out "${TOP_DIR}"/... >./$file_input); then
    cat ./$file_input
    echo '****** go test cases error! ******'
    exit 1
  else
    gocov convert cov.out > gocov.json
    gocov convert cov.out | gocov-html > "$file_detail_output"
    gotestsum --junitfile unit-tests.xml -- -race -gcflags=all=-l "${TOP_DIR}"/...
    exit 0
  fi
}

export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

file_input='testNodeD.txt'
file_detail_output='api.html'

echo "clean old version test results"
if [ -d "${TOP_DIR}"/test ]; then
  rm -rf "${TOP_DIR}"/test
fi
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/

if [ -f "$file_input" ]; then
  rm -rf $file_input
fi
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi

echo "************************************* Start LLT Test *************************************"
execute_test
echo "************************************* End   LLT Test *************************************"