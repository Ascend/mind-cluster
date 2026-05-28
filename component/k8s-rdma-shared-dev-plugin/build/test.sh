#!/bin/bash
# Perform test k8s-rdma-shared-dev-plugin
# Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

# discover packages under pkg/ that contain test files
function discover_test_packages() {
  local test_files
  test_files=$(find "${TOP_DIR}"/pkg -name "*_test.go" -type f 2>/dev/null)
  if [[ -z "$test_files" ]]; then
    echo "no test files found under pkg/"
    exit 0
  fi
  echo "$test_files" | xargs -I{} dirname {} | sort -u | while read -r dir; do
    echo "../${dir#${TOP_DIR}/}"
  done
}

# check third-party tools availability
function check_tools() {
  local missing_tools=()
  for tool in gocov gocov-html gotestsum; do
    if ! command -v "$tool" &>/dev/null; then
      missing_tools+=("$tool")
    fi
  done
  if [ ${#missing_tools[@]} -gt 0 ]; then
    echo "warning: the following tools are not found in PATH: ${missing_tools[*]}"
    echo "warning: coverage html report and junit xml generation will be skipped"
    echo "warning: install them via: go install github.com/axw/gocov/gocov@latest && go install github.com/matm/gocov-html/cmd/gocov-html@latest && go install gotest.tools/gotestsum@latest"
    return 1
  fi
  return 0
}

# execute go test and echo result to report files
function execute_test() {
  local test_packages
  test_packages=$(discover_test_packages)
  echo "test packages: $test_packages"

  if ! (go test -mod=mod -gcflags=all=-l -v -coverprofile cov.out $test_packages >./$file_input); then
    cat ./$file_input
    echo '****** go test cases error! ******'
    exit 1
  else
    if check_tools; then
      gotestsum --junitfile unit-tests.xml -- -mod=mod -gcflags=all=-l -v -coverprofile cov.out $test_packages >./$file_input
      gocov convert cov.out | gocov-html > "$file_detail_output"
    fi

    total_coverage=$(go tool cover -func=cov.out | grep "total:" | awk '{print $3}'| sed 's/%//')
    # round up
    coverage=$(echo "$total_coverage" | awk '{if ($1 >= 0) print ($1 == int($1)) ? int($1) : int($1) + 1;\
                                          else print ($1 == int($1)) ? int($1) : int($1)}')
    if [[ $coverage -ge 1 ]]; then
      echo "coverage passed: $coverage%"
      exit 0
    else
      echo "coverage failed: $coverage%, it needs to be greater than 1%."
      exit 1
    fi
  fi
}

export GO111MODULE="on"
if [ -z "$GOPATH" ]; then
  export GOPATH
  GOPATH=$(go env GOPATH)
fi
export PATH=$GOPATH/bin:$PATH
CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

file_input='testRdmaSharedDp.txt'
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
