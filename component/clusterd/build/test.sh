#!/bin/bash

# Perform test clusterd
# Copyright(C) Huawei Technologies Co.,Ltd. 2024. All rights reserved.
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

CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
export GOTOOLCHAIN="go1.21.13"
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

function filter_cov_by_tested_pkgs() {
  local tested_pkgs
  tested_pkgs=$(go list -buildvcs=false -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' "${TOP_DIR}"/pkg/...)
  awk -v pkgs="$tested_pkgs" '
    NR==1 {print; next}
    {
      file=$1; sub(/:[0-9].*/, "", file)
      n=split(file, p, "/"); pkg=""
      for(i=1;i<n;i++) pkg=pkg p[i]"/"; sub(/\/$/,"", pkg)
      found=0; split(pkgs, arr, "\n"); for(k in arr) {if(arr[k]==pkg){found=1;break}}
      if (found) print
    }
  ' cov.out > cov_filtered.out
}

function execute_test() {
  if ! gotestsum --junitfile unit-tests.xml --jsonfile test.jsonl \
    -- -mod=mod -count=1 -gcflags=all=-l -v -coverprofile cov.out "${TOP_DIR}"/pkg/...; then
    echo '****** go test cases error! ******'
    exit 1
  fi

  filter_cov_by_tested_pkgs

  gocov convert cov.out | gocov-html > "$file_detail_output"
  total_coverage_before_filtered=$(go tool cover -func=cov.out | grep "total:" | awk '{print $3}'| sed 's/%//')
  total_coverage=$(go tool cover -func=cov_filtered.out | grep "total:" | awk '{print $3}'| sed 's/%//')
  # round up
  coverage=$(echo "$total_coverage" | awk '{if ($1 >= 0) print ($1 == int($1)) ? int($1) : int($1) + 1;\
                                        else print ($1 == int($1)) ? int($1) : int($1)}')
  if [[ $coverage -ge 80 ]]; then
    echo "coverage passed: $coverage%"
    exit 0
  else
    echo "coverage failed: $coverage%, it needs to be greater than 80%."
    exit 1
  fi
}

file_detail_output='api.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi
execute_test
echo "************************************* End LLT Test *************************************"
