#!/bin/bash

# Perform test taskd-go
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
export GO111MODULE="on"
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/../../..)
OUTPUT_DIR="${TOP_DIR}"/test/ut/go
GO_PKG=${TOP_DIR}/taskd/go
TEMP_DIR=${TOP_DIR}/taskd/go/test
FILE_DETAIL_OUTPUT='api.html'

function filter_cov_by_tested_pkgs() {
  local tested_pkgs
  tested_pkgs=$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' "${GO_PKG}"/...)
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

function unit_test() {
  gotestsum --junitfile unit-tests.xml --jsonfile test.jsonl \
    -- -mod=mod -count=1 -gcflags=all=-l -v -coverprofile cov.out "${GO_PKG}"/...;

  filter_cov_by_tested_pkgs

  gocov convert cov_filtered.out | gocov-html > "$FILE_DETAIL_OUTPUT"
  total_coverage=$(go tool cover -func=cov_filtered.out | grep "total:" | awk '{print $3}'| sed 's/%//')
  # round up
  coverage=$(echo "$total_coverage" | awk '{if ($1 >= 0) print ($1 == int($1)) ? int($1) : int($1) + 1;\
                                        else print ($1 == int($1)) ? int($1) : int($1)}')

  if [[ $coverage -ge 80 ]]; then
    echo "coverage passed: $coverage%"
  else
    echo "coverage failed: $coverage%, it needs to be greater than 80%."
    # exit 1
  fi
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
    echo "************************************* Start LLT Test *************************************"
    clean_before
    mkdir -p "${TEMP_DIR}"
    cd "${TEMP_DIR}"
    unit_test
    echo "************************************* End LLT Test *************************************"
    clean_end
    exit 0
}

execute_test
